package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/odinnordico/privutil/internal/api"
	pb "github.com/odinnordico/privutil/proto"
)

// The tools below all follow the same shape: a typed input struct (which the SDK
// turns into a documented JSON schema) and a handler that calls the API server,
// surfaces any in-band error via errResult, and returns the response as generic
// structured output (see structured).

// ── Data & formats ──────────────────────────────────────────────────────────
func registerDataTools(srv *mcp.Server, s *api.Server) {
	type convertIn struct {
		Data         string `json:"data" jsonschema:"the document to convert"`
		SourceFormat string `json:"source_format" jsonschema:"source format: json, yaml, xml, toml, or csv"`
		TargetFormat string `json:"target_format" jsonschema:"target format: json, yaml, xml, toml, or csv"`
		CsvDelimiter string `json:"csv_delimiter,omitempty" jsonschema:"single-character CSV delimiter (default ,)"`
		CsvNoHeader  bool   `json:"csv_no_header,omitempty" jsonschema:"treat CSV rows as arrays instead of objects"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "convert", Description: "Convert a document between JSON, YAML, XML, TOML, and CSV."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in convertIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.Convert(ctx, &pb.ConvertRequest{
				Data: in.Data, SourceFormat: dataFormat(in.SourceFormat), TargetFormat: dataFormat(in.TargetFormat),
				CsvDelimiter: in.CsvDelimiter, CsvNoHeader: in.CsvNoHeader,
			})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type validateIn struct {
		Data   string `json:"data" jsonschema:"the document to validate"`
		Format string `json:"format" jsonschema:"format: json, yaml, xml, or toml"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "validate_data", Description: "Validate the syntax of a JSON, YAML, XML, or TOML document."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in validateIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.ValidateData(ctx, &pb.ValidateRequest{Data: in.Data, Format: dataFormat(in.Format)})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type jsonToGoIn struct {
		JSON       string `json:"json" jsonschema:"the JSON to convert"`
		StructName string `json:"struct_name,omitempty" jsonschema:"name for the top-level Go struct"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "json_to_go", Description: "Generate Go struct definitions from a JSON document."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in jsonToGoIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.JsonToGo(ctx, &pb.JsonToGoRequest{Json: in.JSON, StructName: in.StructName})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type sqlIn struct {
		Query string `json:"query" jsonschema:"the SQL to format"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "sql_format", Description: "Pretty-print a SQL query."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in sqlIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.SqlFormat(ctx, &pb.SqlRequest{Query: in.Query})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type textIn struct {
		Text string `json:"text" jsonschema:"the input text"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "markdown_to_html", Description: "Render Markdown to HTML (raw HTML is stripped)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.MarkdownToHtml(ctx, &pb.TextRequest{Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})
	mcp.AddTool(srv, &mcp.Tool{Name: "html_to_markdown", Description: "Convert HTML to Markdown."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.HtmlToMarkdown(ctx, &pb.TextRequest{Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type caseIn struct {
		Text string `json:"text" jsonschema:"the text to convert"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "case_convert", Description: "Convert text into camelCase, snake_case, PascalCase, kebab-case, CONSTANT_CASE, and Title Case."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in caseIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.CaseConvert(ctx, &pb.CaseRequest{Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type escapeIn struct {
		Text   string `json:"text" jsonschema:"the text to escape or unescape"`
		Mode   string `json:"mode" jsonschema:"escaping mode: json, java, sql, or html_entity"`
		Action string `json:"action" jsonschema:"escape or unescape"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "string_escape", Description: "Escape or unescape a string for JSON, Java, SQL, or HTML entities."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in escapeIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.StringEscape(ctx, &pb.EscapeRequest{Text: in.Text, Mode: in.Mode, Action: in.Action})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type listIn struct {
		Text            string `json:"text" jsonschema:"newline-separated list"`
		Action          string `json:"action" jsonschema:"one of: sort_az, sort_za, sort_numeric, shuffle, dedupe, unique_only, duplicates, frequency, reverse, trim, remove_empty"`
		CaseInsensitive bool   `json:"case_insensitive,omitempty" jsonschema:"case-insensitive comparisons"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "list_process", Description: "Process a newline-separated list: sort, dedupe, shuffle, frequency, and more."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.ListProcess(ctx, &pb.ListRequest{Text: in.Text, Action: listAction(in.Action), CaseInsensitive: in.CaseInsensitive})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})
}

// ── Text ────────────────────────────────────────────────────────────────────
func registerTextTools(srv *mcp.Server, s *api.Server) {
	type textIn struct {
		Text string `json:"text" jsonschema:"the input text"`
	}

	mcp.AddTool(srv, &mcp.Tool{Name: "text_inspect", Description: "Count characters, words, lines, and bytes of text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.TextInspect(ctx, &pb.TextInspectRequest{Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type manipIn struct {
		Text   string `json:"text" jsonschema:"the input text (lines)"`
		Action string `json:"action" jsonschema:"one of: sort_az, sort_za, sort_numeric, reverse, dedupe, remove_empty, trim"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "text_manipulate", Description: "Sort, reverse, dedupe, trim, or remove empty lines from text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in manipIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.TextManipulate(ctx, &pb.TextManipulateRequest{Text: in.Text, Action: textAction(in.Action)})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type replaceIn struct {
		Text            string `json:"text" jsonschema:"the input text"`
		Find            string `json:"find" jsonschema:"the substring or regex to find"`
		ReplaceWith     string `json:"replace_with" jsonschema:"the replacement"`
		UseRegex        bool   `json:"use_regex,omitempty" jsonschema:"treat find as a regular expression"`
		CaseInsensitive bool   `json:"case_insensitive,omitempty" jsonschema:"case-insensitive matching"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "text_replace", Description: "Find-and-replace within text (literal or regex)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in replaceIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.TextReplace(ctx, &pb.TextReplaceRequest{
				Text: in.Text, Find: in.Find, ReplaceWith: in.ReplaceWith, UseRegex: in.UseRegex, CaseInsensitive: in.CaseInsensitive,
			})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type regexIn struct {
		Pattern string `json:"pattern" jsonschema:"the RE2 regular expression"`
		Text    string `json:"text" jsonschema:"the text to test"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "regex_test", Description: "Test a regular expression against text and return matches."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in regexIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.RegexTest(ctx, &pb.RegexRequest{Pattern: in.Pattern, Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	mcp.AddTool(srv, &mcp.Tool{Name: "hidden_chars", Description: "Detect hidden/invisible/zero-width characters in text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.HiddenChars(ctx, &pb.HiddenCharsRequest{Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type simIn struct {
		Text1 string `json:"text1" jsonschema:"the first text"`
		Text2 string `json:"text2" jsonschema:"the second text"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "text_similarity", Description: "Levenshtein edit distance and similarity ratio between two texts."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in simIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.TextSimilarity(ctx, &pb.SimilarityRequest{Text1: in.Text1, Text2: in.Text2})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type tokenIn struct {
		Text     string `json:"text" jsonschema:"the text to tokenize"`
		Strategy string `json:"strategy,omitempty" jsonschema:"optional tokenizer id filter (empty returns all)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "token_count", Description: "Count LLM tokens for text across tokenizers (OpenAI BPE + heuristics)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in tokenIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: in.Text, Strategy: in.Strategy})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type spellIn struct {
		Text     string `json:"text" jsonschema:"the text to check"`
		Language string `json:"language,omitempty" jsonschema:"language code: en (default) or es"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "spell_check", Description: "Offline spelling and grammar check (English or Spanish); returns located issues."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in spellIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.SpellCheck(ctx, &pb.SpellCheckRequest{Text: in.Text, Language: in.Language})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	mcp.AddTool(srv, &mcp.Tool{Name: "html_encode", Description: "Escape text into HTML entities."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.HtmlEncode(ctx, &pb.TextRequest{Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})
	mcp.AddTool(srv, &mcp.Tool{Name: "html_decode", Description: "Decode HTML entities back into text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in textIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.HtmlDecode(ctx, &pb.TextRequest{Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})
}

// ── Encoding & ciphers ──────────────────────────────────────────────────────
func registerEncodingTools(srv *mcp.Server, s *api.Server) {
	type encIn struct {
		Text   string `json:"text" jsonschema:"the input text"`
		Action string `json:"action" jsonschema:"encode or decode"`
		Format string `json:"format" jsonschema:"binary, hex, octal, or decimal"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "text_encode", Description: "Encode/decode text as binary, hex, octal, or decimal code points."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in encIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.TextEncode(ctx, &pb.TextEncodeRequest{Text: in.Text, Action: in.Action, Format: in.Format})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type actionText struct {
		Text   string `json:"text" jsonschema:"the input text"`
		Action string `json:"action" jsonschema:"encode or decode"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "morse_code", Description: "Encode/decode Morse code."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in actionText) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.MorseCode(ctx, &pb.MorseRequest{Text: in.Text, Action: in.Action})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})
	mcp.AddTool(srv, &mcp.Tool{Name: "nato_phonetic", Description: "Encode/decode the NATO phonetic alphabet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in actionText) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.NatoAlphabet(ctx, &pb.NatoRequest{Text: in.Text, Action: in.Action})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type caesarIn struct {
		Text   string `json:"text" jsonschema:"the input text"`
		Shift  int32  `json:"shift" jsonschema:"shift 1-25 (ROT13 = 13)"`
		Action string `json:"action" jsonschema:"encode or decode"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "caesar_cipher", Description: "Apply a Caesar/ROT cipher to text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in caesarIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.CaesarCipher(ctx, &pb.CaesarRequest{Text: in.Text, Shift: in.Shift, Action: in.Action})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type obfuscateIn struct {
		Text      string `json:"text" jsonschema:"the text to mask"`
		KeepStart int32  `json:"keep_start,omitempty" jsonschema:"characters to reveal at the start"`
		KeepEnd   int32  `json:"keep_end,omitempty" jsonschema:"characters to reveal at the end"`
		MaskChar  string `json:"mask_char,omitempty" jsonschema:"masking character (default *)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "string_obfuscate", Description: "Mask the middle of a string, keeping a few characters at each end."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in obfuscateIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.StringObfuscate(ctx, &pb.StringObfuscateRequest{Text: in.Text, KeepStart: in.KeepStart, KeepEnd: in.KeepEnd, MaskChar: in.MaskChar})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type numeronymIn struct {
		Text string `json:"text" jsonschema:"the text (e.g. internationalization -> i18n)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "numeronym", Description: "Generate numeronyms (i18n-style abbreviations) for words."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in numeronymIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.NumeronymGenerate(ctx, &pb.NumeronymRequest{Text: in.Text})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})
}

// ── Network & web ───────────────────────────────────────────────────────────
func registerNetworkTools(srv *mcp.Server, s *api.Server) {
	type ipIn struct {
		CIDR string `json:"cidr" jsonschema:"an IPv4/IPv6 CIDR, e.g. 192.168.1.0/24"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "ip_calc", Description: "Subnet calculator: network, broadcast, netmask, host range, and count for a CIDR."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in ipIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.IpCalc(ctx, &pb.IpRequest{Cidr: in.CIDR})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type ipv4In struct {
		Input string `json:"input" jsonschema:"an IPv4 as dotted, decimal, hex, or binary"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "ipv4_convert", Description: "Convert an IPv4 address between dotted, decimal, hex, and binary."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in ipv4In) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.Ipv4Convert(ctx, &pb.Ipv4ConvertRequest{Input: in.Input})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type rangeIn struct {
		Start string `json:"start" jsonschema:"start IPv4 (dotted)"`
		End   string `json:"end" jsonschema:"end IPv4 (dotted)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "ipv4_range_expand", Description: "Expand an IPv4 range into addresses (capped) and a CIDR summary."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in rangeIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.Ipv4RangeExpand(ctx, &pb.Ipv4RangeRequest{Start: in.Start, End: in.End})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type urlIn struct {
		URL string `json:"url" jsonschema:"the URL to parse"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "url_parse", Description: "Parse a URL into scheme, host, port, path, query params, and fragment."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in urlIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.UrlParse(ctx, &pb.UrlParseRequest{Url: in.URL})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type uaIn struct {
		UserAgent string `json:"user_agent" jsonschema:"the User-Agent string"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "user_agent_parse", Description: "Parse a User-Agent string into browser, OS, engine, and device."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in uaIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.UserAgentParse(ctx, &pb.UserAgentParseRequest{UserAgent: in.UserAgent})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type mimeIn struct {
		Query string `json:"query" jsonschema:"a file extension, MIME type, or partial match"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "mime_lookup", Description: "Look up MIME types by extension, type, or partial match."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in mimeIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.MimeLookup(ctx, &pb.MimeLookupRequest{Query: in.Query})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type httpIn struct {
		Query    string `json:"query,omitempty" jsonschema:"status code number or name keyword (empty for all)"`
		Category string `json:"category,omitempty" jsonschema:"1xx, 2xx, 3xx, 4xx, 5xx, or empty for all"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "http_status_search", Description: "Look up HTTP status codes by number, name, or category."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in httpIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.HttpStatusSearch(ctx, &pb.HttpStatusSearchRequest{Query: in.Query, Category: in.Category})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type chmodIn struct {
		Input string `json:"input" jsonschema:"octal (e.g. 755) or symbolic (e.g. rwxr-xr-x) permissions"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "chmod_calc", Description: "Convert file permissions between octal and symbolic, with a breakdown."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in chmodIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.ChmodCalc(ctx, &pb.ChmodRequest{Input: in.Input})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})
}

// ── Math, units & dates ─────────────────────────────────────────────────────
func registerDateMathTools(srv *mcp.Server, s *api.Server) {
	type mathIn struct {
		Expression string `json:"expression" jsonschema:"the math expression to evaluate"`
		Precision  int32  `json:"precision,omitempty" jsonschema:"decimal places 0-15 (default 10)"`
		Degrees    bool   `json:"degrees,omitempty" jsonschema:"trig in degrees instead of radians"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "math_eval", Description: "Evaluate a math expression (functions, constants, trig)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in mathIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.MathEval(ctx, &pb.MathEvalRequest{Expression: in.Expression, Precision: in.Precision, Degrees: in.Degrees})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type pctIn struct {
		Mode string  `json:"mode" jsonschema:"one of: x_of_y, what, change, reverse"`
		A    float64 `json:"a" jsonschema:"first operand"`
		B    float64 `json:"b" jsonschema:"second operand"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "percentage_calc", Description: "Percentage calculations: X% of Y, A is what % of B, % change, and reverse."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in pctIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.PercentageCalc(ctx, &pb.PercentageRequest{Mode: percentMode(in.Mode), A: in.A, B: in.B})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type tempIn struct {
		Value    float64 `json:"value" jsonschema:"the temperature value"`
		FromUnit string  `json:"from_unit" jsonschema:"c, f, or k"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "temp_convert", Description: "Convert a temperature between Celsius, Fahrenheit, and Kelvin."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in tempIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.TempConvert(ctx, &pb.TempConvertRequest{Value: in.Value, FromUnit: in.FromUnit})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type unitIn struct {
		Value    float64 `json:"value" jsonschema:"the value to convert"`
		FromUnit string  `json:"from_unit" jsonschema:"the source unit (depends on category)"`
		Category string  `json:"category" jsonschema:"one of: bytes, length, mass, area, volume, speed"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "unit_convert", Description: "Convert a value across units in a category (bytes, length, mass, area, volume, speed)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in unitIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.UnitConvert(ctx, &pb.UnitConvertRequest{Value: in.Value, FromUnit: in.FromUnit, Category: unitCategory(in.Category)})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type dateDiffIn struct {
		FromDate string `json:"from_date" jsonschema:"start date/datetime (any parseable format)"`
		ToDate   string `json:"to_date" jsonschema:"end date/datetime"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "date_diff", Description: "Calendar-aware difference between two dates."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in dateDiffIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.DateDiff(ctx, &pb.DateDiffRequest{FromDate: in.FromDate, ToDate: in.ToDate})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type leapIn struct {
		Input string `json:"input" jsonschema:"a year, comma-separated years, or a YYYY-YYYY range"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "leap_year", Description: "Check whether years are leap years."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in leapIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.LeapYear(ctx, &pb.LeapYearRequest{Input: in.Input})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type dateAddIn struct {
		Date    string `json:"date" jsonschema:"the base date"`
		Years   int32  `json:"years,omitempty"`
		Months  int32  `json:"months,omitempty"`
		Weeks   int32  `json:"weeks,omitempty"`
		Days    int32  `json:"days,omitempty"`
		Hours   int32  `json:"hours,omitempty"`
		Minutes int32  `json:"minutes,omitempty"`
		Seconds int32  `json:"seconds,omitempty"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "date_add", Description: "Add a duration to a date and describe the result."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in dateAddIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.DateAdd(ctx, &pb.DateAddRequest{
				Date: in.Date, Years: in.Years, Months: in.Months, Weeks: in.Weeks,
				Days: in.Days, Hours: in.Hours, Minutes: in.Minutes, Seconds: in.Seconds,
			})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type dateFmtIn struct {
		DateStr  string `json:"date_str" jsonschema:"the date to reformat"`
		Timezone string `json:"timezone,omitempty" jsonschema:"IANA timezone (default UTC)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "date_format", Description: "Render a date in many common formats for a timezone."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in dateFmtIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.DateFormat(ctx, &pb.DateFormatRequest{DateStr: in.DateStr, Timezone: in.Timezone})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type dateInfoIn struct {
		Date string `json:"date" jsonschema:"the date to inspect"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "date_info", Description: "Facts about a date: weekday, ISO week, quarter, day-of-year, zodiac, season, and more."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in dateInfoIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.DateInfo(ctx, &pb.DateInfoRequest{Date: in.Date})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type timeIn struct {
		Input string `json:"input" jsonschema:"a Unix timestamp or an ISO date string"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "time_convert", Description: "Convert between Unix timestamps and human-readable UTC/local/ISO times."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in timeIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.TimeConvert(ctx, &pb.TimeRequest{Input: in.Input})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})
}

// ── Generators & auth ───────────────────────────────────────────────────────
func registerGeneratorTools(srv *mcp.Server, s *api.Server) {
	type pwIn struct {
		Length    int32  `json:"length,omitempty" jsonschema:"password length"`
		Uppercase bool   `json:"uppercase,omitempty"`
		Lowercase bool   `json:"lowercase,omitempty"`
		Numbers   bool   `json:"numbers,omitempty"`
		Symbols   bool   `json:"symbols,omitempty"`
		Count     int32  `json:"count,omitempty" jsonschema:"how many passwords"`
		Custom    string `json:"custom_chars,omitempty" jsonschema:"custom character set"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "generate_password", Description: "Generate random passwords with configurable character classes."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in pwIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.GeneratePassword(ctx, &pb.PasswordRequest{
				Length: in.Length, Uppercase: in.Uppercase, Lowercase: in.Lowercase,
				Numbers: in.Numbers, Symbols: in.Symbols, Count: in.Count, CustomChars: in.Custom,
			})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type portIn struct {
		Count         int32 `json:"count,omitempty"`
		Min           int32 `json:"min,omitempty"`
		Max           int32 `json:"max,omitempty"`
		ExcludeSystem bool  `json:"exclude_system,omitempty" jsonschema:"exclude ports 0-1023"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "generate_port", Description: "Generate random TCP/UDP port numbers."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in portIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.GeneratePort(ctx, &pb.PortRequest{Count: in.Count, Min: in.Min, Max: in.Max, ExcludeSystem: in.ExcludeSystem})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type macIn struct {
		Count     int32  `json:"count,omitempty"`
		Separator string `json:"separator,omitempty" jsonschema:"':', '-', '.', or '' (none)"`
		Uppercase bool   `json:"uppercase,omitempty"`
		OUI       string `json:"oui,omitempty" jsonschema:"optional 3-byte OUI prefix"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "generate_mac", Description: "Generate random MAC addresses."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in macIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.GenerateMac(ctx, &pb.MacRequest{Count: in.Count, Separator: in.Separator, Uppercase: in.Uppercase, Oui: in.OUI})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type loremIn struct {
		Type  string `json:"type,omitempty" jsonschema:"word, sentence, or paragraph"`
		Count int32  `json:"count,omitempty"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "generate_lorem", Description: "Generate lorem ipsum placeholder text."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in loremIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.GenerateLorem(ctx, &pb.LoremRequest{Type: in.Type, Count: in.Count})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type ulidIn struct {
		Count     int32 `json:"count,omitempty"`
		Monotonic bool  `json:"monotonic,omitempty" jsonschema:"monotonic ordering within the same millisecond"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "ulid_generate", Description: "Generate ULIDs (sortable unique identifiers)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in ulidIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.UlidGenerate(ctx, &pb.UlidRequest{Count: in.Count, Monotonic: in.Monotonic})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type rsaIn struct {
		Bits int32 `json:"bits,omitempty" jsonschema:"key size in bits (1024-8192, default 2048)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "generate_rsa_keypair", Description: "Generate an RSA key pair (PEM PKCS#1 private + PKIX public)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in rsaIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.GenerateRsaKeyPair(ctx, &pb.RsaKeyRequest{Bits: in.Bits})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type otpGenIn struct {
		Secret         string `json:"secret,omitempty" jsonschema:"base32 secret (omit with generate_secret)"`
		Type           string `json:"type,omitempty" jsonschema:"totp (default) or hotp"`
		Counter        int64  `json:"counter,omitempty" jsonschema:"HOTP counter"`
		Digits         int32  `json:"digits,omitempty" jsonschema:"6 (default) or 8"`
		Period         int32  `json:"period,omitempty" jsonschema:"TOTP period seconds (default 30)"`
		Algo           string `json:"algo,omitempty" jsonschema:"sha1 (default), sha256, or sha512"`
		GenerateSecret bool   `json:"generate_secret,omitempty" jsonschema:"generate a new random secret"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "otp_generate", Description: "Generate a TOTP/HOTP code (and optionally a new secret)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in otpGenIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.OtpGenerate(ctx, &pb.OtpRequest{
				Secret: in.Secret, Type: in.Type, Counter: in.Counter, Digits: in.Digits,
				Period: in.Period, Algo: in.Algo, GenerateSecret: in.GenerateSecret,
			})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type otpValIn struct {
		Secret string `json:"secret" jsonschema:"base32 secret"`
		Code   string `json:"code" jsonschema:"the code to validate"`
		Window int32  `json:"window,omitempty" jsonschema:"allowed drift in periods (default 1)"`
		Period int32  `json:"period,omitempty" jsonschema:"period seconds (default 30)"`
		Algo   string `json:"algo,omitempty" jsonschema:"sha1 (default), sha256, or sha512"`
		Digits int32  `json:"digits,omitempty" jsonschema:"6 (default) or 8"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "otp_validate", Description: "Validate a TOTP/HOTP code against a secret."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in otpValIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.OtpValidate(ctx, &pb.OtpValidateRequest{
				Secret: in.Secret, Code: in.Code, Window: in.Window, Period: in.Period, Algo: in.Algo, Digits: in.Digits,
			})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type basicAuthIn struct {
		Username string `json:"username" jsonschema:"the username"`
		Password string `json:"password" jsonschema:"the password"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "basic_auth_generate", Description: "Build an HTTP Basic Authorization header from a username and password."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in basicAuthIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.BasicAuthGenerate(ctx, &pb.BasicAuthRequest{Username: in.Username, Password: in.Password})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})
}

// ── DevOps & misc ───────────────────────────────────────────────────────────
func registerMiscTools(srv *mcp.Server, s *api.Server) {
	type cronIn struct {
		Expression string `json:"expression" jsonschema:"the cron expression"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "cron_explain", Description: "Explain a cron expression in plain English and show upcoming runs."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in cronIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.CronExplain(ctx, &pb.CronRequest{Expression: in.Expression})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type dockerIn struct {
		Command string `json:"command" jsonschema:"a full 'docker run ...' command"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "docker_run_to_compose", Description: "Convert a 'docker run' command into a docker-compose service."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in dockerIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.DockerRunToCompose(ctx, &pb.DockerRunToComposeRequest{Command: in.Command})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type gitIn struct {
		Query    string `json:"query,omitempty" jsonschema:"search term (empty for all)"`
		Category string `json:"category,omitempty" jsonschema:"filter by category (empty for all)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "git_cheat_sheet", Description: "Search a git command cheat sheet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in gitIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.GitCheatSheet(ctx, &pb.GitCheatSheetRequest{Query: in.Query, Category: in.Category})
			if err != nil {
				return nil, nil, err
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type certIn struct {
		Data string `json:"data" jsonschema:"a PEM-encoded X.509 certificate"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "cert_parse", Description: "Parse a PEM X.509 certificate (subject, issuer, validity, SANs)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in certIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.CertParse(ctx, &pb.CertRequest{Data: in.Data})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type colorIn struct {
		Input string `json:"input" jsonschema:"a color as #hex or rgb(...)"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "color_convert", Description: "Convert a color to HEX, RGB, and HSL."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in colorIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.ColorConvert(ctx, &pb.ColorRequest{Input: in.Input})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})

	type svgIn struct {
		SVG    string `json:"svg" jsonschema:"the SVG markup"`
		Preset string `json:"preset,omitempty" jsonschema:"safe (default), aggressive, or minimal"`
	}
	mcp.AddTool(srv, &mcp.Tool{Name: "svg_optimize", Description: "Optimize/minify SVG markup."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in svgIn) (*mcp.CallToolResult, map[string]any, error) {
			resp, err := s.SvgOptimize(ctx, &pb.SvgOptimizeRequest{Svg: in.SVG, Preset: in.Preset})
			if err != nil {
				return nil, nil, err
			}
			if resp.Error != "" {
				return errResult(resp.Error), nil, nil
			}
			out, err := structured(resp)
			return nil, out, err
		})
}
