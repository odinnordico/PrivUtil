package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"strconv"
	"strings"

	pb "github.com/odinnordico/privutil/proto"
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
