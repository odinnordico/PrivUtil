package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"math/big"
	"net/url"
	"strconv"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	pb "github.com/odinnordico/privutil/proto"
	"github.com/yuin/goldmark"
)

func (s *Server) Base64Encode(ctx context.Context, req *pb.Base64Request) (*pb.Base64Response, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(req.Text))
	return &pb.Base64Response{
		Text: encoded,
	}, nil
}

func (s *Server) Base64Decode(ctx context.Context, req *pb.Base64Request) (*pb.Base64Response, error) {
	decoded, err := base64.StdEncoding.DecodeString(req.Text)
	if err != nil {
		return &pb.Base64Response{
			Error: fmt.Sprintf("Failed to decode: %v", err),
		}, nil
	}
	return &pb.Base64Response{
		Text: string(decoded),
	}, nil
}

func (s *Server) UrlEncode(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	return &pb.TextResponse{Text: url.QueryEscape(req.Text)}, nil
}

func (s *Server) UrlDecode(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	decoded, err := url.QueryUnescape(req.Text)
	if err != nil {
		return &pb.TextResponse{Text: "Error: " + err.Error()}, nil
	}
	return &pb.TextResponse{Text: decoded}, nil
}

func (s *Server) HtmlEncode(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	return &pb.TextResponse{Text: html.EscapeString(req.Text)}, nil
}

func (s *Server) HtmlDecode(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	return &pb.TextResponse{Text: html.UnescapeString(req.Text)}, nil
}

func (s *Server) StringEscape(ctx context.Context, req *pb.EscapeRequest) (*pb.EscapeResponse, error) {
	text := req.Text
	var res string
	var err error

	switch req.Mode {
	case "json":
		if req.Action == "escape" {
			b, _ := jsonMarshal(text)
			res = string(b)
		} else {
			if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
				if err := jsonUnmarshal([]byte(text), &res); err != nil {
					return &pb.EscapeResponse{Error: "Invalid JSON string"}, nil
				}
			} else {
				if err := jsonUnmarshal([]byte("\""+text+"\""), &res); err != nil {
					return &pb.EscapeResponse{Error: "Could not unescape"}, nil
				}
			}
		}
	case "html_entity":
		if req.Action == "escape" {
			res = html.EscapeString(text)
		} else {
			res = html.UnescapeString(text)
		}
	case "url":
		if req.Action == "escape" {
			res = url.QueryEscape(text)
		} else {
			res, err = url.QueryUnescape(text)
		}
	case "sql":
		if req.Action == "escape" {
			res = strings.ReplaceAll(text, "'", "''")
		} else {
			res = strings.ReplaceAll(text, "''", "'")
		}
	case "java":
		if req.Action == "escape" {
			res = strconv.Quote(text)
		} else {
			res, err = strconv.Unquote(text)
		}
	default:
		return &pb.EscapeResponse{Error: "Unknown mode"}, nil
	}

	if err != nil {
		return &pb.EscapeResponse{Error: err.Error()}, nil
	}

	return &pb.EscapeResponse{Result: res}, nil
}

func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s *Server) BaseConvert(ctx context.Context, req *pb.BaseConvertRequest) (*pb.BaseConvertResponse, error) {
	input := strings.TrimSpace(req.Input)
	sourceBase := int(req.SourceBase)

	if input == "" {
		return &pb.BaseConvertResponse{Error: "Input cannot be empty"}, nil
	}

	// Strip prefixes
	inputLower := strings.ToLower(input)
	if sourceBase == 16 && strings.HasPrefix(inputLower, "0x") {
		input = input[2:]
	} else if sourceBase == 8 && strings.HasPrefix(inputLower, "0o") {
		input = input[2:]
	} else if sourceBase == 2 && strings.HasPrefix(inputLower, "0b") {
		input = input[2:]
	}

	num := new(big.Int)

	if sourceBase == 64 {
		// Custom parsing for Base64 integer
		const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
		for _, r := range input {
			idx := strings.IndexRune(alphabet, r)
			if idx == -1 {
				return &pb.BaseConvertResponse{Error: "Invalid base64 character"}, nil
			}
			num.Mul(num, big.NewInt(64))
			num.Add(num, big.NewInt(int64(idx)))
		}
	} else {
		_, ok := num.SetString(input, sourceBase)
		if !ok {
			return &pb.BaseConvertResponse{Error: fmt.Sprintf("Invalid input for base %d", sourceBase)}, nil
		}
	}

	// Format targets
	decimalStr := num.Text(10)
	hexStr := strings.ToUpper(num.Text(16))
	binaryStr := num.Text(2)
	octalStr := num.Text(8)

	// Base64 formatter
	sign := num.Sign()
	if sign == 0 {
		return &pb.BaseConvertResponse{
			Decimal: "0",
			Hex:     "0",
			Binary:  "0",
			Octal:   "0",
			Base64:  "A",
		}, nil
	}

	absNum := new(big.Int).Abs(num)
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var b64Builder strings.Builder
	zero := big.NewInt(0)
	sixtyFour := big.NewInt(64)
	
	for absNum.Cmp(zero) > 0 {
		mod := new(big.Int)
		absNum.DivMod(absNum, sixtyFour, mod)
		b64Builder.WriteByte(alphabet[mod.Int64()])
	}
	
	// Reverse the builder since we extracted least significant digits first
	b64Bytes := []byte(b64Builder.String())
	for i, j := 0, len(b64Bytes)-1; i < j; i, j = i+1, j-1 {
		b64Bytes[i], b64Bytes[j] = b64Bytes[j], b64Bytes[i]
	}
	base64Str := string(b64Bytes)
	
	if sign < 0 {
		// Just prefixing minus based on math magnitude base conversion assumption
		decimalStr = "-" + decimalStr
		hexStr = "-" + hexStr
		binaryStr = "-" + binaryStr
		octalStr = "-" + octalStr
		base64Str = "-" + base64Str
	}

	return &pb.BaseConvertResponse{
		Decimal: decimalStr,
		Hex:     hexStr,
		Binary:  binaryStr,
		Octal:   octalStr,
		Base64:  base64Str,
	}, nil
}

func (s *Server) MarkdownToHtml(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(req.Text), &buf); err != nil {
		return &pb.TextResponse{Text: "Error: " + err.Error()}, nil
	}
	return &pb.TextResponse{Text: buf.String()}, nil
}

func (s *Server) HtmlToMarkdown(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	markdown, err := htmltomarkdown.ConvertString(req.Text)
	if err != nil {
		return &pb.TextResponse{Text: "Error: " + err.Error()}, nil
	}
	return &pb.TextResponse{Text: markdown}, nil
}
