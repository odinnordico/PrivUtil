package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/odinnordico/privutil/internal/api"
	pb "github.com/odinnordico/privutil/proto"
)

// registerTools wires a curated set of PrivUtil RPCs as MCP tools. Each handler
// unmarshals typed input (schema inferred from the struct + jsonschema tags),
// calls the API server, and returns typed structured output. Add more by
// following the same pattern.
func registerTools(srv *mcp.Server, s *api.Server) {
	// ── Hashing / crypto ──────────────────────────────────────────────────────
	type hashIn struct {
		Text string `json:"text" jsonschema:"the text to hash"`
		Algo string `json:"algo" jsonschema:"algorithm: md5, sha1, sha256 (default), sha512, or bcrypt"`
		Cost *int32 `json:"cost,omitempty" jsonschema:"bcrypt cost 4-15 (only used when algo is bcrypt)"`
	}
	type hashOut struct {
		Hash string `json:"hash"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "calculate_hash", Description: "Compute a cryptographic hash (md5, sha1, sha256, sha512, or bcrypt) of text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in hashIn) (*mcp.CallToolResult, hashOut, error) {
			resp, err := s.CalculateHash(ctx, &pb.HashRequest{Text: in.Text, Algo: in.Algo, Cost: in.Cost})
			if err != nil {
				return nil, hashOut{}, err
			}
			return nil, hashOut{Hash: resp.Hash}, nil
		})

	type hmacIn struct {
		Message string `json:"message" jsonschema:"the message to authenticate"`
		Secret  string `json:"secret" jsonschema:"the shared secret key"`
		Algo    string `json:"algo" jsonschema:"algorithm: sha256 (default), sha512, sha1, or md5"`
	}
	type hmacOut struct {
		Hex    string `json:"hex"`
		Base64 string `json:"base64"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "hmac_generate", Description: "Generate an HMAC of a message with a secret key."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in hmacIn) (*mcp.CallToolResult, hmacOut, error) {
			resp, err := s.HmacGenerate(ctx, &pb.HmacRequest{Message: in.Message, Secret: in.Secret, Algo: in.Algo})
			if err != nil {
				return nil, hmacOut{}, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), hmacOut{}, nil
			}
			return nil, hmacOut{Hex: resp.Hex, Base64: resp.Base64}, nil
		})

	// ── Generators ────────────────────────────────────────────────────────────
	type uuidIn struct {
		Version   string `json:"version,omitempty" jsonschema:"UUID version: v1, v2, v3, v4 (default), v5, v6, v7, v8"`
		Count     int32  `json:"count,omitempty" jsonschema:"how many to generate (1-100, default 1)"`
		Hyphen    bool   `json:"hyphen,omitempty" jsonschema:"include hyphen separators (default false)"`
		Uppercase bool   `json:"uppercase,omitempty" jsonschema:"uppercase output (default false)"`
		Namespace string `json:"namespace,omitempty" jsonschema:"for v3/v5/v8: dns, url, oid, or x500"`
	}
	type uuidOut struct {
		UUIDs []string `json:"uuids"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "generate_uuid", Description: "Generate one or more UUIDs (versions 1-8)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in uuidIn) (*mcp.CallToolResult, uuidOut, error) {
			resp, err := s.GenerateUuid(ctx, &pb.UuidRequest{
				Version: in.Version, Count: in.Count, Hyphen: in.Hyphen, Uppercase: in.Uppercase, Namespace: in.Namespace,
			})
			if err != nil {
				return nil, uuidOut{}, err
			}
			return nil, uuidOut{UUIDs: resp.Uuids}, nil
		})

	// ── Encoding ──────────────────────────────────────────────────────────────
	type textIn struct {
		Text string `json:"text" jsonschema:"the input text"`
	}

	type b64EncOut struct {
		Encoded string `json:"encoded"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "base64_encode", Description: "Base64-encode UTF-8 text (standard encoding)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, b64EncOut, error) {
			resp, err := s.Base64Encode(ctx, &pb.Base64Request{Text: in.Text})
			if err != nil {
				return nil, b64EncOut{}, err
			}
			return nil, b64EncOut{Encoded: resp.Text}, nil
		})

	type b64DecOut struct {
		Text     string `json:"text" jsonschema:"the decoded bytes as text (may be garbled if binary)"`
		MimeType string `json:"mime_type" jsonschema:"detected content type of the decoded data"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "base64_decode", Description: "Decode a Base64 string (accepts standard/URL-safe, padded or not) and report the detected content type."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, b64DecOut, error) {
			resp, err := s.Base64Decode(ctx, &pb.Base64Request{Text: in.Text})
			if err != nil {
				return nil, b64DecOut{}, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), b64DecOut{}, nil
			}
			return nil, b64DecOut{Text: string(resp.Data), MimeType: resp.MimeType}, nil
		})

	type encOut struct {
		Result string `json:"result"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "url_encode", Description: "URL/percent-encode text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, encOut, error) {
			resp, err := s.UrlEncode(ctx, &pb.TextRequest{Text: in.Text})
			if err != nil {
				return nil, encOut{}, err
			}
			return nil, encOut{Result: resp.Text}, nil
		})
	mcp.AddTool(srv, &mcp.Tool{Name: "url_decode", Description: "Decode URL/percent-encoded text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, encOut, error) {
			resp, err := s.UrlDecode(ctx, &pb.TextRequest{Text: in.Text})
			if err != nil {
				return nil, encOut{}, err
			}
			return nil, encOut{Result: resp.Text}, nil
		})

	type baseConvIn struct {
		Input      string `json:"input" jsonschema:"the number as a string"`
		SourceBase int32  `json:"source_base" jsonschema:"base of the input: 2, 8, 10, 16, or 64"`
	}
	type baseConvOut struct {
		Decimal string `json:"decimal"`
		Hex     string `json:"hex"`
		Binary  string `json:"binary"`
		Octal   string `json:"octal"`
		Base64  string `json:"base64"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "base_convert", Description: "Convert a number between bases (2, 8, 10, 16, 64)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in baseConvIn) (*mcp.CallToolResult, baseConvOut, error) {
			resp, err := s.BaseConvert(ctx, &pb.BaseConvertRequest{Input: in.Input, SourceBase: in.SourceBase})
			if err != nil {
				return nil, baseConvOut{}, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), baseConvOut{}, nil
			}
			return nil, baseConvOut{Decimal: resp.Decimal, Hex: resp.Hex, Binary: resp.Binary, Octal: resp.Octal, Base64: resp.Base64}, nil
		})

	// ── Data / text ───────────────────────────────────────────────────────────
	type jsonIn struct {
		Text     string `json:"text" jsonschema:"the JSON to format"`
		Indent   string `json:"indent,omitempty" jsonschema:"indentation: 2 (default), 4, tab, or min to minify"`
		SortKeys bool   `json:"sort_keys,omitempty" jsonschema:"sort object keys alphabetically"`
	}
	type jsonOut struct {
		Formatted string `json:"formatted"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "json_format", Description: "Pretty-print, minify, or sort-keys a JSON document."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in jsonIn) (*mcp.CallToolResult, jsonOut, error) {
			resp, err := s.JsonFormat(ctx, &pb.JsonFormatRequest{Text: in.Text, Indent: in.Indent, SortKeys: in.SortKeys})
			if err != nil {
				return nil, jsonOut{}, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), jsonOut{}, nil
			}
			return nil, jsonOut{Formatted: resp.Text}, nil
		})

	type jwtIn struct {
		Token string `json:"token" jsonschema:"the JWT to decode (not verified)"`
	}
	type jwtOut struct {
		Header  string `json:"header"`
		Payload string `json:"payload"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "jwt_decode", Description: "Decode a JWT's header and payload (signature is NOT verified)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in jwtIn) (*mcp.CallToolResult, jwtOut, error) {
			resp, err := s.JwtDecode(ctx, &pb.JwtRequest{Token: in.Token})
			if err != nil {
				return nil, jwtOut{}, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), jwtOut{}, nil
			}
			return nil, jwtOut{Header: resp.Header, Payload: resp.Payload}, nil
		})

	type slugIn struct {
		Text      string `json:"text" jsonschema:"the text to slugify"`
		Separator string `json:"separator,omitempty" jsonschema:"separator: - (default), _, or empty"`
		Uppercase bool   `json:"uppercase,omitempty" jsonschema:"uppercase output (default lowercase)"`
		MaxLen    int32  `json:"max_len,omitempty" jsonschema:"maximum length (0 = unlimited)"`
	}
	type slugOut struct {
		Slug string `json:"slug"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "slugify", Description: "Turn text into a URL-safe slug."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in slugIn) (*mcp.CallToolResult, slugOut, error) {
			resp, err := s.Slugify(ctx, &pb.SlugifyRequest{Text: in.Text, Separator: in.Separator, Uppercase: in.Uppercase, MaxLen: in.MaxLen})
			if err != nil {
				return nil, slugOut{}, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), slugOut{}, nil
			}
			return nil, slugOut{Slug: resp.Result}, nil
		})

	type diffIn struct {
		Text1 string `json:"text1" jsonschema:"the original text"`
		Text2 string `json:"text2" jsonschema:"the modified text"`
	}
	type diffOut struct {
		DiffHTML string `json:"diff_html" jsonschema:"inline diff as HTML (ins/del/span spans)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "text_diff", Description: "Compute an inline (unified) diff of two texts, returned as HTML."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in diffIn) (*mcp.CallToolResult, diffOut, error) {
			resp, err := s.Diff(ctx, &pb.DiffRequest{Text1: in.Text1, Text2: in.Text2})
			if err != nil {
				return nil, diffOut{}, err
			}
			return nil, diffOut{DiffHTML: resp.DiffHtml}, nil
		})
}
