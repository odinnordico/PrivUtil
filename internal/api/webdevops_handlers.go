package api

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	pb "github.com/odinnordico/privutil/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── URL Parser ───────────────────────────────────────────────────────────────

func (s *Server) UrlParse(_ context.Context, req *pb.UrlParseRequest) (*pb.UrlParseResponse, error) {
	raw := strings.TrimSpace(req.Url)
	if raw == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	// If no scheme, add https:// to allow net/url to parse host correctly
	parseTarget := raw
	addedScheme := false
	if !strings.Contains(raw, "://") {
		parseTarget = "https://" + raw
		addedScheme = true
	}

	u, err := url.Parse(parseTarget)
	if err != nil {
		return &pb.UrlParseResponse{IsValid: false, Error: err.Error()}, nil
	}

	scheme := u.Scheme
	if addedScheme {
		scheme = ""
	}

	var username, password string
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	hostname := u.Hostname()
	port := u.Port()
	host := hostname
	if port != "" {
		host = hostname + ":" + port
	}

	var params []*pb.QueryParam
	for k, vals := range u.Query() {
		for _, v := range vals {
			params = append(params, &pb.QueryParam{Key: k, Value: v})
		}
	}
	sort.Slice(params, func(i, j int) bool { return params[i].Key < params[j].Key })

	// Normalize: rebuild URL without extra spaces
	normalized := u.String()
	if addedScheme {
		normalized = strings.TrimPrefix(normalized, "https://")
	}

	return &pb.UrlParseResponse{
		Scheme:      scheme,
		Username:    username,
		Password:    password,
		Host:        host,
		Hostname:    hostname,
		Port:        port,
		Path:        u.Path,
		Query:       u.RawQuery,
		QueryParams: params,
		Fragment:    u.Fragment,
		Normalized:  normalized,
		IsValid:     true,
	}, nil
}

// ─── User-Agent Parser ────────────────────────────────────────────────────────

var (
	uaBotPatterns = regexp.MustCompile(`(?i)(bot|crawler|spider|scraper|slurp|facebookexternalhit|linkedinbot|twitterbot|whatsapp|telegrambot|discordbot|pingdom|datadog|monitoring|googlebot|bingbot|yandexbot|baiduspider|duckduckbot|sogou|exabot|ia_archiver)`)

	uaBrowserPatterns = []struct {
		name    string
		re      *regexp.Regexp
		version *regexp.Regexp
	}{
		{"Samsung Internet", regexp.MustCompile(`SamsungBrowser`), regexp.MustCompile(`SamsungBrowser/([\d.]+)`)},
		{"Edge", regexp.MustCompile(`Edg(?:e|A|iOS)?/`), regexp.MustCompile(`Edg(?:e|A|iOS)?/([\d.]+)`)},
		{"Opera", regexp.MustCompile(`OPR/|Opera/`), regexp.MustCompile(`(?:OPR|Opera)/([\d.]+)`)},
		{"Brave", regexp.MustCompile(`Brave`), regexp.MustCompile(`Brave/([\d.]+)`)},
		{"Vivaldi", regexp.MustCompile(`Vivaldi/`), regexp.MustCompile(`Vivaldi/([\d.]+)`)},
		{"Chrome", regexp.MustCompile(`(?:Chrome|CriOS)/`), regexp.MustCompile(`(?:Chrome|CriOS)/([\d.]+)`)},
		{"Firefox", regexp.MustCompile(`(?:Firefox|FxiOS)/`), regexp.MustCompile(`(?:Firefox|FxiOS)/([\d.]+)`)},
		{"Safari", regexp.MustCompile(`Safari/`), regexp.MustCompile(`Version/([\d.]+)`)},
		{"IE", regexp.MustCompile(`(?:MSIE\s|Trident/)`), regexp.MustCompile(`(?:MSIE\s([\d.]+)|rv:([\d.]+))`)},
	}

	uaOSPatterns = []struct {
		name    string
		re      *regexp.Regexp
		version *regexp.Regexp
	}{
		{"Android", regexp.MustCompile(`Android`), regexp.MustCompile(`Android\s+([\d.]+)`)},
		{"iOS", regexp.MustCompile(`(?:iPhone|iPad|iPod)`), regexp.MustCompile(`OS\s+([\d_]+)`)},
		{"Windows", regexp.MustCompile(`Windows NT`), regexp.MustCompile(`Windows NT\s*([\d.]+)`)},
		{"macOS", regexp.MustCompile(`Mac OS X`), regexp.MustCompile(`Mac OS X\s*([\d_.]+)`)},
		{"ChromeOS", regexp.MustCompile(`CrOS`), regexp.MustCompile(`CrOS\s+\S+\s+([\d.]+)`)},
		{"Linux", regexp.MustCompile(`Linux`), nil},
	}
)

func uaWindowsVersion(v string) string {
	switch v {
	case "10.0":
		return "10/11"
	case "6.3":
		return "8.1"
	case "6.2":
		return "8"
	case "6.1":
		return "7"
	case "6.0":
		return "Vista"
	case "5.2":
		return "Server 2003 / XP x64"
	case "5.1":
		return "XP"
	default:
		return v
	}
}

func (s *Server) UserAgentParse(_ context.Context, req *pb.UserAgentParseRequest) (*pb.UserAgentParseResponse, error) {
	ua := strings.TrimSpace(req.UserAgent)
	if ua == "" {
		return nil, status.Error(codes.InvalidArgument, "user_agent is required")
	}

	res := &pb.UserAgentParseResponse{}

	// Bot detection
	if uaBotPatterns.MatchString(ua) {
		res.IsBot = true
		res.DeviceType = "bot"
		if m := regexp.MustCompile(`(?i)([\w.]+[Bb]ot|[Ss]pider|[Cc]rawler)`).FindString(ua); m != "" {
			res.BrowserName = m
		} else {
			res.BrowserName = "Bot/Crawler"
		}
	}

	// Browser detection (ordered: most specific first)
	if !res.IsBot {
		for _, bp := range uaBrowserPatterns {
			if bp.re.MatchString(ua) {
				res.BrowserName = bp.name
				if m := bp.version.FindStringSubmatch(ua); len(m) > 1 {
					for _, g := range m[1:] {
						if g != "" {
							res.BrowserVersion = g
							break
						}
					}
				}
				break
			}
		}
	}

	// OS detection
	for _, op := range uaOSPatterns {
		if op.re.MatchString(ua) {
			res.OsName = op.name
			if op.version != nil {
				if m := op.version.FindStringSubmatch(ua); len(m) > 1 && m[1] != "" {
					v := strings.ReplaceAll(m[1], "_", ".")
					if res.OsName == "Windows" {
						v = uaWindowsVersion(v)
					}
					res.OsVersion = v
				}
			}
			break
		}
	}

	// Device type
	if !res.IsBot {
		switch {
		case regexp.MustCompile(`(?i)iPad|Tablet`).MatchString(ua):
			res.DeviceType = "tablet"
			res.IsMobile = true
		case regexp.MustCompile(`(?i)Mobi|Mobile|iPhone|iPod|Android.*(?:Mobile)`).MatchString(ua):
			res.DeviceType = "mobile"
			res.IsMobile = true
		default:
			res.DeviceType = "desktop"
		}
	}

	// Engine detection
	switch res.BrowserName {
	case "Firefox":
		res.Engine = "Gecko"
		if m := regexp.MustCompile(`Gecko/([\d.]+)`).FindStringSubmatch(ua); len(m) > 1 {
			res.EngineVersion = m[1]
		}
	case "Safari":
		res.Engine = "WebKit"
		if m := regexp.MustCompile(`WebKit/([\d.]+)`).FindStringSubmatch(ua); len(m) > 1 {
			res.EngineVersion = m[1]
		}
	case "IE":
		res.Engine = "Trident"
		if m := regexp.MustCompile(`Trident/([\d.]+)`).FindStringSubmatch(ua); len(m) > 1 {
			res.EngineVersion = m[1]
		}
	default:
		if strings.Contains(ua, "AppleWebKit") {
			res.Engine = "Blink"
			if m := regexp.MustCompile(`AppleWebKit/([\d.]+)`).FindStringSubmatch(ua); len(m) > 1 {
				res.EngineVersion = m[1]
			}
		}
	}

	// Build display fields
	addField := func(label, value string) {
		if value != "" {
			res.Fields = append(res.Fields, &pb.UAParsedField{Label: label, Value: value})
		}
	}
	addField("Browser", res.BrowserName+" "+res.BrowserVersion)
	addField("OS", res.OsName+" "+res.OsVersion)
	addField("Device", res.DeviceType)
	addField("Engine", res.Engine+" "+res.EngineVersion)
	if res.IsBot {
		res.Fields = append(res.Fields, &pb.UAParsedField{Label: "Bot", Value: "yes"})
	}

	return res, nil
}

// ─── HTTP Status Codes ────────────────────────────────────────────────────────

type httpStatusDef struct {
	code        int32
	name        string
	description string
}

var httpStatuses = []httpStatusDef{
	// 1xx
	{100, "Continue", "The server has received the request headers and the client should proceed to send the request body."},
	{101, "Switching Protocols", "The requester has asked the server to switch protocols and the server has agreed to do so."},
	{102, "Processing", "A WebDAV request may contain many sub-requests involving file operations, requiring a long time to complete."},
	{103, "Early Hints", "Primarily intended to be used with the Link header, letting the user agent preload resources."},
	// 2xx
	{200, "OK", "The request has succeeded. The meaning depends on the HTTP method used."},
	{201, "Created", "The request has been fulfilled and has resulted in one or more new resources being created."},
	{202, "Accepted", "The request has been accepted for processing, but the processing has not been completed."},
	{203, "Non-Authoritative Information", "The returned metadata is not exactly the same as available from the origin server."},
	{204, "No Content", "The server has successfully fulfilled the request and there is no additional content to send."},
	{205, "Reset Content", "The server has fulfilled the request and desires that the user agent reset the document view."},
	{206, "Partial Content", "The server is delivering only part of the resource due to a range header sent by the client."},
	{207, "Multi-Status", "The message body contains XML and can contain multiple separate response codes (WebDAV)."},
	{208, "Already Reported", "The members of a DAV binding have already been enumerated in a preceding part of the response."},
	{226, "IM Used", "The server has fulfilled a GET request for the resource using one or more instance manipulations."},
	// 3xx
	{300, "Multiple Choices", "The request has more than one possible response. The user agent or user should choose one."},
	{301, "Moved Permanently", "The URL of the requested resource has been changed permanently. The new URL is given in the response."},
	{302, "Found", "The URI of the requested resource has been temporarily changed. Future requests should use the current URI."},
	{303, "See Other", "The server sent this response to direct the client to get the requested resource at another URI via GET."},
	{304, "Not Modified", "This response tells the client that the response has not been modified, so the client can use its cached version."},
	{305, "Use Proxy", "The requested response must be accessed by a proxy. This response code is deprecated."},
	{307, "Temporary Redirect", "The server sends this response to direct the client to get the requested resource at another URI using the same method."},
	{308, "Permanent Redirect", "The resource is now permanently located at another URI using the same HTTP method."},
	// 4xx
	{400, "Bad Request", "The server could not understand the request due to invalid syntax or missing required fields."},
	{401, "Unauthorized", "Authentication is required and has failed or has not yet been provided."},
	{402, "Payment Required", "Reserved for future use. Some services use this for rate limiting or subscription requirements."},
	{403, "Forbidden", "The client does not have access rights to the content. Unlike 401, the client's identity is known."},
	{404, "Not Found", "The server can not find the requested resource. The URL is not recognized."},
	{405, "Method Not Allowed", "The request method is known by the server but is not supported by the target resource."},
	{406, "Not Acceptable", "The server cannot produce a response matching the list of acceptable values defined in the request's headers."},
	{407, "Proxy Authentication Required", "Authentication is required to be performed by a proxy."},
	{408, "Request Timeout", "This response is sent when the server wants to shut down an idle connection."},
	{409, "Conflict", "The request conflicts with the current state of the server."},
	{410, "Gone", "The content has been permanently deleted from the server, with no forwarding address."},
	{411, "Length Required", "Server rejected the request because the Content-Length header field is not defined and the server requires it."},
	{412, "Precondition Failed", "The client has indicated preconditions in its headers which the server does not meet."},
	{413, "Content Too Large", "Request entity is larger than limits defined by server; the server might close the connection."},
	{414, "URI Too Long", "The URI requested by the client is longer than the server is willing to interpret."},
	{415, "Unsupported Media Type", "The media format of the requested data is not supported by the server."},
	{416, "Range Not Satisfiable", "The range specified by the Range header field in the request can't be fulfilled."},
	{417, "Expectation Failed", "The expectation indicated by the Expect request header field can't be met by the server."},
	{418, "I'm a teapot", "The server refuses to brew coffee because it is, permanently, a teapot. (RFC 2324, April Fools)"},
	{421, "Misdirected Request", "The request was directed at a server that is not able to produce a response."},
	{422, "Unprocessable Content", "The request was well-formed but was unable to be followed due to semantic errors."},
	{423, "Locked", "The resource that is being accessed is locked (WebDAV)."},
	{424, "Failed Dependency", "The request failed because it depended on another request and that request failed (WebDAV)."},
	{425, "Too Early", "Indicates that the server is unwilling to risk processing a request that might be replayed."},
	{426, "Upgrade Required", "The server refuses to perform the request using the current protocol."},
	{428, "Precondition Required", "The origin server requires the request to be conditional."},
	{429, "Too Many Requests", "The user has sent too many requests in a given amount of time (rate limiting)."},
	{431, "Request Header Fields Too Large", "The server is unwilling to process the request because its header fields are too large."},
	{451, "Unavailable For Legal Reasons", "The user requested a resource that cannot legally be provided (DMCA, government censorship, etc)."},
	// 5xx
	{500, "Internal Server Error", "The server has encountered a situation it doesn't know how to handle."},
	{501, "Not Implemented", "The request method is not supported by the server and cannot be handled."},
	{502, "Bad Gateway", "The server, while acting as a gateway, received an invalid response from an upstream server."},
	{503, "Service Unavailable", "The server is not ready to handle the request; commonly due to maintenance or overload."},
	{504, "Gateway Timeout", "The server, while acting as a gateway, could not get a response in time from an upstream server."},
	{505, "HTTP Version Not Supported", "The HTTP version used in the request is not supported by the server."},
	{506, "Variant Also Negotiates", "The server has an internal configuration error: the chosen variant resource is itself engaged in content negotiation."},
	{507, "Insufficient Storage", "The method could not be performed on the resource because the server is unable to store the representation needed (WebDAV)."},
	{508, "Loop Detected", "The server detected an infinite loop while processing the request (WebDAV)."},
	{510, "Not Extended", "Further extensions to the request are required for the server to fulfil it."},
	{511, "Network Authentication Required", "The client needs to authenticate to gain network access."},
}

func (s *Server) HttpStatusSearch(_ context.Context, req *pb.HttpStatusSearchRequest) (*pb.HttpStatusSearchResponse, error) {
	q := strings.ToLower(strings.TrimSpace(req.Query))
	cat := strings.ToLower(strings.TrimSpace(req.Category))

	var entries []*pb.HttpStatusEntry
	for _, hs := range httpStatuses {
		category := fmt.Sprintf("%dxx", hs.code/100)

		if cat != "" && cat != category {
			continue
		}
		if q != "" {
			codeStr := strconv.Itoa(int(hs.code))
			if !strings.Contains(codeStr, q) &&
				!strings.Contains(strings.ToLower(hs.name), q) &&
				!strings.Contains(strings.ToLower(hs.description), q) {
				continue
			}
		}
		entries = append(entries, &pb.HttpStatusEntry{
			Code:        hs.code,
			Name:        hs.name,
			Description: hs.description,
			Category:    category,
		})
	}
	return &pb.HttpStatusSearchResponse{Entries: entries}, nil
}

// ─── MIME Type Lookup ─────────────────────────────────────────────────────────

type mimeDef struct {
	mimeType    string
	extensions  string
	category    string
	description string
}

var mimeDatabase = []mimeDef{
	// Text
	{"text/plain", "txt, text, conf, def, list, log", "Text", "Plain text document"},
	{"text/html", "html, htm, shtml", "Text", "HyperText Markup Language"},
	{"text/css", "css", "Text", "Cascading Style Sheets"},
	{"text/javascript", "js, mjs", "Text", "JavaScript source code"},
	{"text/csv", "csv", "Text", "Comma-Separated Values"},
	{"text/markdown", "md, markdown", "Text", "Markdown document"},
	{"text/xml", "xml", "Text", "Extensible Markup Language"},
	{"text/calendar", "ics, ifb", "Text", "iCalendar format"},
	{"text/rtf", "rtf", "Text", "Rich Text Format"},
	{"text/vcard", "vcf, vcard", "Text", "vCard contact data"},
	{"text/tab-separated-values", "tsv", "Text", "Tab-Separated Values"},
	// Image
	{"image/jpeg", "jpg, jpeg, jpe", "Image", "JPEG image"},
	{"image/png", "png", "Image", "Portable Network Graphics"},
	{"image/gif", "gif", "Image", "Graphics Interchange Format"},
	{"image/webp", "webp", "Image", "WebP image"},
	{"image/svg+xml", "svg, svgz", "Image", "Scalable Vector Graphics"},
	{"image/avif", "avif", "Image", "AV1 Image File Format"},
	{"image/bmp", "bmp, dib", "Image", "Bitmap image"},
	{"image/tiff", "tif, tiff", "Image", "Tagged Image File Format"},
	{"image/x-icon", "ico", "Image", "Windows icon"},
	{"image/heic", "heic, heif", "Image", "High Efficiency Image Format"},
	{"image/apng", "apng", "Image", "Animated Portable Network Graphics"},
	// Audio
	{"audio/mpeg", "mp3, mpga", "Audio", "MPEG audio (MP3)"},
	{"audio/ogg", "ogg, oga", "Audio", "OGG audio"},
	{"audio/wav", "wav, wave", "Audio", "Waveform Audio File Format"},
	{"audio/flac", "flac", "Audio", "Free Lossless Audio Codec"},
	{"audio/aac", "aac, adts", "Audio", "Advanced Audio Coding"},
	{"audio/webm", "weba", "Audio", "WebM audio"},
	{"audio/midi", "mid, midi, kar, rmi", "Audio", "MIDI audio"},
	{"audio/opus", "opus", "Audio", "Opus audio"},
	{"audio/x-m4a", "m4a", "Audio", "MPEG-4 audio"},
	// Video
	{"video/mp4", "mp4, m4v, mp4v", "Video", "MPEG-4 video"},
	{"video/webm", "webm", "Video", "WebM video"},
	{"video/ogg", "ogv", "Video", "OGG video"},
	{"video/quicktime", "mov, qt", "Video", "QuickTime video"},
	{"video/x-msvideo", "avi", "Video", "Audio Video Interleave"},
	{"video/x-matroska", "mkv, mk3d, mks", "Video", "Matroska video"},
	{"video/x-flv", "flv", "Video", "Flash Video"},
	{"video/mpeg", "mpeg, mpg, mpe, m1v, m2v", "Video", "MPEG video"},
	{"video/3gpp", "3gp", "Video", "3GPP video"},
	// Application
	{"application/json", "json", "Application", "JSON data"},
	{"application/ld+json", "jsonld", "Application", "JSON-LD linked data"},
	{"application/xml", "xml, xsl, xsd, rng", "Application", "XML data"},
	{"application/zip", "zip", "Application", "ZIP archive"},
	{"application/pdf", "pdf", "Application", "Portable Document Format"},
	{"application/octet-stream", "bin, exe, dll, so, dmg, iso", "Application", "Binary/arbitrary data"},
	{"application/gzip", "gz, gzip", "Application", "Gzip compressed archive"},
	{"application/x-bzip2", "bz2, boz", "Application", "Bzip2 archive"},
	{"application/x-tar", "tar", "Application", "TAR archive"},
	{"application/x-7z-compressed", "7z", "Application", "7-Zip archive"},
	{"application/x-rar-compressed", "rar", "Application", "RAR archive"},
	{"application/x-xz", "xz", "Application", "XZ compressed data"},
	{"application/msword", "doc, dot", "Application", "Microsoft Word document"},
	{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "docx", "Application", "Microsoft Word (OpenXML)"},
	{"application/vnd.ms-excel", "xls, xlm, xla, xlc, xlt, xlw", "Application", "Microsoft Excel spreadsheet"},
	{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "xlsx", "Application", "Microsoft Excel (OpenXML)"},
	{"application/vnd.ms-powerpoint", "ppt, pps, pot", "Application", "Microsoft PowerPoint presentation"},
	{"application/vnd.openxmlformats-officedocument.presentationml.presentation", "pptx", "Application", "Microsoft PowerPoint (OpenXML)"},
	{"application/wasm", "wasm", "Application", "WebAssembly binary"},
	{"application/x-www-form-urlencoded", "", "Application", "HTML form data (URL encoded)"},
	{"multipart/form-data", "", "Application", "HTML form data with file uploads"},
	{"application/x-httpd-php", "php, phtml", "Application", "PHP source code"},
	{"application/x-sh", "sh", "Application", "Shell script"},
	{"application/x-csh", "csh", "Application", "C shell script"},
	{"application/toml", "toml", "Application", "TOML configuration"},
	{"application/yaml", "yaml, yml", "Application", "YAML data"},
	// Font
	{"font/ttf", "ttf", "Font", "TrueType font"},
	{"font/otf", "otf", "Font", "OpenType font"},
	{"font/woff", "woff", "Font", "Web Open Font Format"},
	{"font/woff2", "woff2", "Font", "Web Open Font Format 2"},
	{"font/collection", "ttc", "Font", "TrueType font collection"},
	// Archive
	{"application/vnd.debian.binary-package", "deb", "Archive", "Debian package"},
	{"application/x-rpm", "rpm", "Archive", "RPM package"},
	{"application/x-apple-diskimage", "dmg", "Archive", "Apple Disk Image"},
	// Special
	{"application/atom+xml", "atom", "Feed", "Atom syndication feed"},
	{"application/rss+xml", "rss", "Feed", "RSS feed"},
	{"application/manifest+json", "webmanifest", "Web", "Web App Manifest"},
	{"text/cache-manifest", "appcache", "Web", "HTML5 App Cache Manifest"},
}

func (s *Server) MimeLookup(_ context.Context, req *pb.MimeLookupRequest) (*pb.MimeLookupResponse, error) {
	q := strings.ToLower(strings.TrimSpace(req.Query))
	if q == "" {
		// Return all
		var entries []*pb.MimeEntry
		for _, m := range mimeDatabase {
			entries = append(entries, &pb.MimeEntry{
				MimeType:    m.mimeType,
				Extensions:  m.extensions,
				Category:    m.category,
				Description: m.description,
			})
		}
		return &pb.MimeLookupResponse{Entries: entries}, nil
	}

	// Clean up extension query
	extQ := strings.TrimPrefix(q, ".")

	var entries []*pb.MimeEntry
	for _, m := range mimeDatabase {
		if strings.Contains(strings.ToLower(m.mimeType), q) ||
			strings.Contains(strings.ToLower(m.description), q) ||
			strings.Contains(strings.ToLower(m.category), q) ||
			containsWord(strings.ToLower(m.extensions), extQ) {
			entries = append(entries, &pb.MimeEntry{
				MimeType:    m.mimeType,
				Extensions:  m.extensions,
				Category:    m.category,
				Description: m.description,
			})
		}
	}
	return &pb.MimeLookupResponse{Entries: entries}, nil
}

func containsWord(s, word string) bool {
	for _, part := range strings.Split(s, ",") {
		if strings.TrimSpace(part) == word {
			return true
		}
	}
	return false
}

// ─── Docker run → Compose ─────────────────────────────────────────────────────

type dockerConfig struct {
	name        string
	image       string
	cmd         []string
	ports       []string
	volumes     []string
	envVars     []string
	envFiles    []string
	restart     string
	network     string
	hostname    string
	user        string
	workdir     string
	entrypoint  string
	memLimit    string
	cpus        string
	cpuShares   string
	labels      []string
	extraHosts  []string
	links       []string
	capAdd      []string
	capDrop     []string
	devices     []string
	dns         []string
	dnsSearch   []string
	shmSize     string
	tmpfs       []string
	securityOpt []string
	logDriver   string
	logOpts     []string
	tty         bool
	stdinOpen   bool
	privileged  bool
	readOnly    bool
	rm          bool
	warnings    []string
}

// tokenizeShell splits a command string respecting single/double quotes and backslash escapes.
func tokenizeShell(s string) []string {
	var tokens []string
	var cur strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for _, r := range s {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && !inSingle {
			escaped = true
			continue
		}
		if r == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if r == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}
		if (r == ' ' || r == '\t' || r == '\n') && !inSingle && !inDouble {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
			continue
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

func parseDockerRun(tokens []string) (*dockerConfig, error) {
	cfg := &dockerConfig{}
	// Skip "docker" and "run" tokens
	i := 0
	for i < len(tokens) && (tokens[i] == "docker" || tokens[i] == "run") {
		i++
	}

	consumeNext := func() (string, bool) {
		if i+1 < len(tokens) && !strings.HasPrefix(tokens[i+1], "-") {
			i++
			return tokens[i], true
		}
		// Value might start with - (e.g. env vars like -e=-SOMETHING)
		return "", false
	}

	// Flags that take a value immediately (--flag=value) or as next token (--flag value)
	takesValue := map[string]bool{
		"--name": true, "-n": true,
		"--publish": true, "-p": true,
		"--volume": true, "-v": true,
		"--env": true, "-e": true,
		"--env-file": true,
		"--restart": true,
		"--network": true, "--net": true,
		"--hostname": true, "-h": true,
		"--user": true, "-u": true,
		"--workdir": true, "-w": true,
		"--entrypoint": true,
		"--memory": true, "-m": true,
		"--cpus": true,
		"--cpu-shares": true,
		"--label": true, "-l": true,
		"--add-host": true,
		"--link": true,
		"--cap-add": true,
		"--cap-drop": true,
		"--device": true,
		"--dns": true,
		"--dns-search": true,
		"--shm-size": true,
		"--tmpfs": true,
		"--security-opt": true,
		"--log-driver": true,
		"--log-opt": true,
		"--platform": true,
		"--pull": true,
	}

	imageFound := false
	for i < len(tokens) {
		tok := tokens[i]

		// Handle --flag=value syntax
		if strings.HasPrefix(tok, "-") && strings.Contains(tok, "=") {
			parts := strings.SplitN(tok, "=", 2)
			tok = parts[0]
			tokens = append(tokens[:i], append([]string{parts[0], parts[1]}, tokens[i+1:]...)...)
		}

		getVal := func() string {
			if i+1 < len(tokens) {
				i++
				return tokens[i]
			}
			return ""
		}

		switch {
		case tok == "--name" || tok == "-n":
			cfg.name = getVal()
		case tok == "--publish" || tok == "-p":
			cfg.ports = append(cfg.ports, getVal())
		case tok == "--volume" || tok == "-v":
			cfg.volumes = append(cfg.volumes, getVal())
		case tok == "--env" || tok == "-e":
			cfg.envVars = append(cfg.envVars, getVal())
		case tok == "--env-file":
			cfg.envFiles = append(cfg.envFiles, getVal())
		case tok == "--restart":
			cfg.restart = getVal()
		case tok == "--network" || tok == "--net":
			cfg.network = getVal()
		case tok == "--hostname" || tok == "-h":
			cfg.hostname = getVal()
		case tok == "--user" || tok == "-u":
			cfg.user = getVal()
		case tok == "--workdir" || tok == "-w":
			cfg.workdir = getVal()
		case tok == "--entrypoint":
			cfg.entrypoint = getVal()
		case tok == "--memory" || tok == "-m":
			cfg.memLimit = getVal()
		case tok == "--cpus":
			cfg.cpus = getVal()
		case tok == "--cpu-shares":
			cfg.cpuShares = getVal()
		case tok == "--label" || tok == "-l":
			cfg.labels = append(cfg.labels, getVal())
		case tok == "--add-host":
			cfg.extraHosts = append(cfg.extraHosts, getVal())
		case tok == "--link":
			cfg.links = append(cfg.links, getVal())
		case tok == "--cap-add":
			cfg.capAdd = append(cfg.capAdd, getVal())
		case tok == "--cap-drop":
			cfg.capDrop = append(cfg.capDrop, getVal())
		case tok == "--device":
			cfg.devices = append(cfg.devices, getVal())
		case tok == "--dns":
			cfg.dns = append(cfg.dns, getVal())
		case tok == "--dns-search":
			cfg.dnsSearch = append(cfg.dnsSearch, getVal())
		case tok == "--shm-size":
			cfg.shmSize = getVal()
		case tok == "--tmpfs":
			cfg.tmpfs = append(cfg.tmpfs, getVal())
		case tok == "--security-opt":
			cfg.securityOpt = append(cfg.securityOpt, getVal())
		case tok == "--log-driver":
			cfg.logDriver = getVal()
		case tok == "--log-opt":
			cfg.logOpts = append(cfg.logOpts, getVal())
		case tok == "-d" || tok == "--detach":
			// default in compose, skip
		case tok == "--rm":
			cfg.rm = true
			cfg.warnings = append(cfg.warnings, "--rm: no direct equivalent; consider 'restart: \"no\"' or omit restart policy")
		case tok == "--privileged":
			cfg.privileged = true
		case tok == "--read-only":
			cfg.readOnly = true
		case tok == "--tty" || tok == "-t":
			cfg.tty = true
		case tok == "--interactive" || tok == "-i":
			cfg.stdinOpen = true
		// Combined short flags like -it, -tid, etc.
		case strings.HasPrefix(tok, "-") && !strings.HasPrefix(tok, "--") && len(tok) > 2:
			for _, c := range tok[1:] {
				switch c {
				case 't':
					cfg.tty = true
				case 'i':
					cfg.stdinOpen = true
				case 'd':
					// detach, skip
				}
			}
		// Ignore known no-value flags
		case tok == "--no-healthcheck" || tok == "--disable-content-trust":
			// skip
		// Platform / pull flags (ignored with warning)
		case tok == "--platform":
			val := getVal()
			cfg.warnings = append(cfg.warnings, fmt.Sprintf("--platform=%s: not directly supported in Compose v3; add it as a top-level platform: field if needed", val))
		case tok == "--pull":
			getVal() // consume value
			cfg.warnings = append(cfg.warnings, "--pull: not a Compose field; use 'docker compose pull' or 'build: pull_policy' instead")
		case strings.HasPrefix(tok, "-"):
			// Unknown flag — try to consume value if next token doesn't start with -
			if _, ok := consumeNext(); ok {
				cfg.warnings = append(cfg.warnings, fmt.Sprintf("unknown flag %s (with value) ignored", tok))
			} else {
				cfg.warnings = append(cfg.warnings, fmt.Sprintf("unknown flag %s ignored", tok))
			}
		default:
			if !imageFound {
				cfg.image = tok
				imageFound = true
			} else {
				cfg.cmd = append(cfg.cmd, tok)
			}
			_ = takesValue
		}
		i++
	}

	if cfg.image == "" {
		return nil, fmt.Errorf("no image specified in docker run command")
	}
	return cfg, nil
}

func yamlStr(v string) string {
	// Quote if contains special YAML characters
	needsQuote := strings.ContainsAny(v, ":{},[]#&*?|-<>=!%@`\"'\\")
	if needsQuote || v == "" {
		return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return v
}

func dockerConfigToCompose(cfg *dockerConfig) string {
	// Determine service name
	svcName := cfg.name
	if svcName == "" {
		// Derive from image: strip tag and registry
		img := cfg.image
		if idx := strings.LastIndex(img, "/"); idx >= 0 {
			img = img[idx+1:]
		}
		if idx := strings.Index(img, ":"); idx >= 0 {
			img = img[:idx]
		}
		svcName = img
		// Replace non-alphanumeric with underscore
		var sb strings.Builder
		for _, r := range svcName {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
				sb.WriteRune(r)
			} else {
				sb.WriteRune('_')
			}
		}
		svcName = sb.String()
		if svcName == "" {
			svcName = "app"
		}
	}

	var b strings.Builder
	w := func(format string, a ...any) {
		fmt.Fprintf(&b, format+"\n", a...)
	}
	ind := func(n int, format string, a ...any) {
		fmt.Fprintf(&b, strings.Repeat("  ", n)+format+"\n", a...)
	}

	w("version: \"3.8\"")
	w("services:")
	ind(1, "%s:", svcName)
	ind(2, "image: %s", yamlStr(cfg.image))

	if cfg.name != "" {
		ind(2, "container_name: %s", yamlStr(cfg.name))
	}
	if cfg.hostname != "" {
		ind(2, "hostname: %s", yamlStr(cfg.hostname))
	}
	if cfg.restart != "" {
		ind(2, "restart: %s", yamlStr(cfg.restart))
	}
	if len(cfg.ports) > 0 {
		ind(2, "ports:")
		for _, p := range cfg.ports {
			ind(3, "- %s", yamlStr(p))
		}
	}
	if len(cfg.volumes) > 0 {
		ind(2, "volumes:")
		for _, v := range cfg.volumes {
			ind(3, "- %s", yamlStr(v))
		}
	}
	if len(cfg.envVars) > 0 {
		ind(2, "environment:")
		for _, e := range cfg.envVars {
			ind(3, "- %s", yamlStr(e))
		}
	}
	if len(cfg.envFiles) > 0 {
		ind(2, "env_file:")
		for _, f := range cfg.envFiles {
			ind(3, "- %s", yamlStr(f))
		}
	}
	if cfg.network != "" && cfg.network != "bridge" {
		ind(2, "networks:")
		ind(3, "- %s", yamlStr(cfg.network))
	}
	if cfg.user != "" {
		ind(2, "user: %s", yamlStr(cfg.user))
	}
	if cfg.workdir != "" {
		ind(2, "working_dir: %s", yamlStr(cfg.workdir))
	}
	if cfg.entrypoint != "" {
		ind(2, "entrypoint: %s", yamlStr(cfg.entrypoint))
	}
	if len(cfg.cmd) > 0 {
		ind(2, "command: %s", yamlStr(strings.Join(cfg.cmd, " ")))
	}
	if cfg.memLimit != "" {
		ind(2, "mem_limit: %s", yamlStr(cfg.memLimit))
	}
	if cfg.cpus != "" || cfg.cpuShares != "" {
		ind(2, "deploy:")
		if cfg.cpus != "" {
			ind(3, "resources:")
			ind(4, "limits:")
			ind(5, "cpus: %s", yamlStr(cfg.cpus))
		}
	}
	if cfg.privileged {
		ind(2, "privileged: true")
	}
	if cfg.readOnly {
		ind(2, "read_only: true")
	}
	if cfg.tty {
		ind(2, "tty: true")
	}
	if cfg.stdinOpen {
		ind(2, "stdin_open: true")
	}
	if cfg.shmSize != "" {
		ind(2, "shm_size: %s", yamlStr(cfg.shmSize))
	}
	if len(cfg.labels) > 0 {
		ind(2, "labels:")
		for _, lbl := range cfg.labels {
			ind(3, "- %s", yamlStr(lbl))
		}
	}
	if len(cfg.extraHosts) > 0 {
		ind(2, "extra_hosts:")
		for _, h := range cfg.extraHosts {
			ind(3, "- %s", yamlStr(h))
		}
	}
	if len(cfg.links) > 0 {
		ind(2, "links:")
		for _, ln := range cfg.links {
			ind(3, "- %s", yamlStr(ln))
		}
	}
	if len(cfg.capAdd) > 0 {
		ind(2, "cap_add:")
		for _, c := range cfg.capAdd {
			ind(3, "- %s", c)
		}
	}
	if len(cfg.capDrop) > 0 {
		ind(2, "cap_drop:")
		for _, c := range cfg.capDrop {
			ind(3, "- %s", c)
		}
	}
	if len(cfg.devices) > 0 {
		ind(2, "devices:")
		for _, d := range cfg.devices {
			ind(3, "- %s", yamlStr(d))
		}
	}
	if len(cfg.dns) > 0 {
		ind(2, "dns:")
		for _, d := range cfg.dns {
			ind(3, "- %s", d)
		}
	}
	if len(cfg.dnsSearch) > 0 {
		ind(2, "dns_search:")
		for _, d := range cfg.dnsSearch {
			ind(3, "- %s", d)
		}
	}
	if len(cfg.tmpfs) > 0 {
		ind(2, "tmpfs:")
		for _, t := range cfg.tmpfs {
			ind(3, "- %s", yamlStr(t))
		}
	}
	if len(cfg.securityOpt) > 0 {
		ind(2, "security_opt:")
		for _, o := range cfg.securityOpt {
			ind(3, "- %s", yamlStr(o))
		}
	}
	if cfg.logDriver != "" {
		ind(2, "logging:")
		ind(3, "driver: %s", yamlStr(cfg.logDriver))
		if len(cfg.logOpts) > 0 {
			ind(3, "options:")
			for _, o := range cfg.logOpts {
				parts := strings.SplitN(o, "=", 2)
				if len(parts) == 2 {
					ind(4, "%s: %s", yamlStr(parts[0]), yamlStr(parts[1]))
				} else {
					ind(4, "%s: \"\"", yamlStr(o))
				}
			}
		}
	}

	// Named network definition if custom network was used
	if cfg.network != "" && cfg.network != "bridge" && cfg.network != "host" && cfg.network != "none" {
		w("")
		w("networks:")
		ind(1, "%s:", yamlStr(cfg.network))
		ind(2, "external: true")
	}

	return strings.TrimRight(b.String(), "\n")
}

func (s *Server) DockerRunToCompose(_ context.Context, req *pb.DockerRunToComposeRequest) (*pb.DockerRunToComposeResponse, error) {
	cmd := strings.TrimSpace(req.Command)
	if cmd == "" {
		return nil, status.Error(codes.InvalidArgument, "command is required")
	}

	tokens := tokenizeShell(cmd)
	cfg, err := parseDockerRun(tokens)
	if err != nil {
		return &pb.DockerRunToComposeResponse{Error: err.Error()}, nil
	}

	yaml := dockerConfigToCompose(cfg)
	svcName := cfg.name
	if svcName == "" {
		svcName = cfg.image
	}

	return &pb.DockerRunToComposeResponse{
		ComposeYaml: yaml,
		Warnings:    cfg.warnings,
		ServiceName: svcName,
		Image:       cfg.image,
	}, nil
}

// ─── Git Cheat Sheet ──────────────────────────────────────────────────────────

type gitCmdDef struct {
	command     string
	description string
	examples    []string
}

type gitCategoryDef struct {
	name     string
	commands []gitCmdDef
}

var gitData = []gitCategoryDef{
	{
		name: "Setup & Config",
		commands: []gitCmdDef{
			{"git config --global user.name \"Your Name\"", "Set your global commit author name", nil},
			{"git config --global user.email \"you@example.com\"", "Set your global commit author email", nil},
			{"git config --list", "List all config settings (local + global)", nil},
			{"git config --list --show-origin", "List config settings with their source file", nil},
			{"git config --global core.editor vim", "Set the default text editor", []string{"git config --global core.editor code --wait"}},
			{"git config --global init.defaultBranch main", "Set default branch name for new repos", nil},
			{"git config --global alias.lg \"log --oneline --graph --all\"", "Create a command alias", nil},
		},
	},
	{
		name: "Creating & Cloning",
		commands: []gitCmdDef{
			{"git init", "Initialize a local repository in the current directory", nil},
			{"git init <directory>", "Create a new repository in the specified directory", nil},
			{"git clone <url>", "Clone a remote repository locally", []string{"git clone https://github.com/user/repo.git"}},
			{"git clone <url> <directory>", "Clone into a specific folder", nil},
			{"git clone --depth 1 <url>", "Shallow clone with only the latest commit (faster)", nil},
			{"git clone --branch <branch> <url>", "Clone a specific branch", nil},
		},
	},
	{
		name: "Staging & Committing",
		commands: []gitCmdDef{
			{"git status", "Show working tree status (staged, unstaged, untracked)", nil},
			{"git status -s", "Short/compact status output", nil},
			{"git add <file>", "Stage a specific file", []string{"git add README.md", "git add src/"}},
			{"git add .", "Stage all changes in the current directory", nil},
			{"git add -A", "Stage all changes including deletions", nil},
			{"git add -p", "Interactively stage hunks of changes", nil},
			{"git commit -m \"message\"", "Commit staged changes with a message", nil},
			{"git commit -am \"message\"", "Stage all tracked files and commit in one step", nil},
			{"git commit --amend", "Amend the last commit (opens editor)", nil},
			{"git commit --amend --no-edit", "Amend last commit without changing the message", nil},
			{"git reset HEAD <file>", "Unstage a file (keep changes in working tree)", nil},
			{"git restore --staged <file>", "Unstage a file (modern syntax)", nil},
			{"git restore <file>", "Discard working directory changes for a file", nil},
			{"git checkout -- <file>", "Discard working directory changes (classic syntax)", nil},
			{"git rm <file>", "Remove a file and stage the deletion", nil},
			{"git rm --cached <file>", "Remove a file from the index only (keep on disk)", nil},
			{"git mv <old> <new>", "Move or rename a file", nil},
		},
	},
	{
		name: "Branching",
		commands: []gitCmdDef{
			{"git branch", "List local branches (current branch marked with *)", nil},
			{"git branch -a", "List all branches including remote-tracking branches", nil},
			{"git branch -v", "List branches with last commit", nil},
			{"git branch <name>", "Create a new branch at the current commit", nil},
			{"git branch -d <name>", "Delete a merged branch", nil},
			{"git branch -D <name>", "Force-delete a branch (even if unmerged)", nil},
			{"git branch -m <old> <new>", "Rename a branch", nil},
			{"git checkout <branch>", "Switch to an existing branch", nil},
			{"git checkout -b <branch>", "Create and switch to a new branch", nil},
			{"git switch <branch>", "Switch to a branch (modern syntax)", nil},
			{"git switch -c <branch>", "Create and switch to a new branch (modern)", nil},
			{"git switch -c <branch> --track origin/<branch>", "Create branch tracking a remote branch", nil},
		},
	},
	{
		name: "Merging & Rebasing",
		commands: []gitCmdDef{
			{"git merge <branch>", "Merge branch into the current branch", nil},
			{"git merge --no-ff <branch>", "Merge without fast-forward (always creates a merge commit)", nil},
			{"git merge --squash <branch>", "Squash all branch commits into one staged change", nil},
			{"git merge --abort", "Abort an in-progress merge", nil},
			{"git cherry-pick <commit>", "Apply a specific commit onto the current branch", []string{"git cherry-pick abc1234"}},
			{"git cherry-pick <from>..<to>", "Apply a range of commits", nil},
			{"git rebase <branch>", "Rebase current branch onto another branch", nil},
			{"git rebase -i HEAD~<n>", "Interactive rebase of the last n commits", []string{"git rebase -i HEAD~3"}},
			{"git rebase --continue", "Continue rebase after resolving conflicts", nil},
			{"git rebase --abort", "Abort the current rebase", nil},
			{"git rebase --skip", "Skip the current commit during rebase", nil},
		},
	},
	{
		name: "Remote Repositories",
		commands: []gitCmdDef{
			{"git remote -v", "List configured remote repositories", nil},
			{"git remote add <name> <url>", "Add a new remote", []string{"git remote add origin https://github.com/user/repo.git"}},
			{"git remote remove <name>", "Remove a remote", nil},
			{"git remote rename <old> <new>", "Rename a remote", nil},
			{"git remote set-url <name> <url>", "Change the URL of a remote", nil},
			{"git fetch", "Fetch all remotes without merging", nil},
			{"git fetch <remote>", "Fetch a specific remote", nil},
			{"git fetch --prune", "Fetch and remove stale remote-tracking branches", nil},
			{"git pull", "Fetch and merge (or rebase) current branch", nil},
			{"git pull --rebase", "Fetch and rebase instead of merge", nil},
			{"git push <remote> <branch>", "Push branch to remote", []string{"git push origin main"}},
			{"git push -u origin <branch>", "Push and set upstream tracking", nil},
			{"git push --force-with-lease", "Force push safely (fails if remote has new commits)", nil},
			{"git push origin --delete <branch>", "Delete a remote branch", nil},
			{"git push --tags", "Push all local tags to remote", nil},
		},
	},
	{
		name: "Inspection & Diff",
		commands: []gitCmdDef{
			{"git log", "Show commit history", nil},
			{"git log --oneline", "Compact one-line commit log", nil},
			{"git log --oneline --graph --all --decorate", "Visual branch graph", nil},
			{"git log -p", "Show commits with patches (diffs)", nil},
			{"git log -n <count>", "Show last n commits", []string{"git log -n 10"}},
			{"git log --author=\"name\"", "Filter commits by author", nil},
			{"git log --since=\"2 weeks ago\"", "Filter commits by date", nil},
			{"git log --grep=\"keyword\"", "Filter commits by message keyword", nil},
			{"git log <file>", "Show commits that changed a specific file", nil},
			{"git diff", "Show unstaged changes", nil},
			{"git diff --staged", "Show staged changes (to be committed)", nil},
			{"git diff <branch1>..<branch2>", "Compare two branches", nil},
			{"git diff HEAD~1 HEAD", "Compare last two commits", nil},
			{"git show <commit>", "Show details of a commit", nil},
			{"git blame <file>", "Show who last modified each line of a file", nil},
			{"git shortlog -sn", "Summary of commits per author", nil},
			{"git describe --tags", "Show the most recent tag reachable from HEAD", nil},
		},
	},
	{
		name: "Stashing",
		commands: []gitCmdDef{
			{"git stash", "Stash current working directory changes", nil},
			{"git stash push -m \"message\"", "Stash with a descriptive message", nil},
			{"git stash push -u", "Stash including untracked files", nil},
			{"git stash list", "List all stashes", nil},
			{"git stash pop", "Apply and remove the latest stash", nil},
			{"git stash apply stash@{n}", "Apply a specific stash without removing it", []string{"git stash apply stash@{0}"}},
			{"git stash drop stash@{n}", "Delete a specific stash", nil},
			{"git stash clear", "Remove all stashes", nil},
			{"git stash branch <branch>", "Create a new branch from a stash", nil},
			{"git stash show -p", "Show the diff of the latest stash", nil},
		},
	},
	{
		name: "Tags",
		commands: []gitCmdDef{
			{"git tag", "List all tags", nil},
			{"git tag -l \"v1.*\"", "List tags matching a pattern", nil},
			{"git tag <name>", "Create a lightweight tag at HEAD", []string{"git tag v1.0.0"}},
			{"git tag -a <name> -m \"msg\"", "Create an annotated tag with a message", []string{"git tag -a v1.0.0 -m \"Release v1.0.0\""}},
			{"git tag -a <name> <commit>", "Tag a specific commit", nil},
			{"git push origin <tag>", "Push a specific tag to remote", nil},
			{"git push origin --tags", "Push all tags to remote", nil},
			{"git tag -d <name>", "Delete a local tag", nil},
			{"git push origin --delete <tag>", "Delete a remote tag", nil},
			{"git checkout <tag>", "Check out a tag (detached HEAD)", nil},
		},
	},
	{
		name: "Undoing & Resetting",
		commands: []gitCmdDef{
			{"git revert <commit>", "Create a new commit that undoes a specific commit", nil},
			{"git revert HEAD", "Revert the last commit", nil},
			{"git revert <from>..<to>", "Revert a range of commits", nil},
			{"git reset --soft HEAD~1", "Undo last commit; keep changes staged", nil},
			{"git reset --mixed HEAD~1", "Undo last commit; keep changes unstaged", nil},
			{"git reset --hard HEAD~1", "Undo last commit; discard all changes", nil},
			{"git reset --hard origin/<branch>", "Reset to match the remote branch exactly", nil},
			{"git clean -fd", "Remove untracked files and directories", nil},
			{"git clean -fdx", "Remove untracked files, directories, and ignored files", nil},
			{"git reflog", "Show the history of HEAD changes (safety net for lost commits)", nil},
			{"git bisect start", "Begin binary search to find the commit that introduced a bug", nil},
			{"git bisect good <commit>", "Mark a commit as bug-free", nil},
			{"git bisect bad <commit>", "Mark a commit as containing the bug", nil},
			{"git bisect reset", "End bisect session", nil},
		},
	},
	{
		name: "Advanced",
		commands: []gitCmdDef{
			{"git worktree add <path> <branch>", "Check out a branch in a new working directory", nil},
			{"git worktree list", "List all working trees", nil},
			{"git submodule add <url>", "Add a submodule", nil},
			{"git submodule update --init --recursive", "Initialize and update all submodules", nil},
			{"git archive --format=zip HEAD > out.zip", "Export the repository as a ZIP file", nil},
			{"git grep <pattern>", "Search the working tree for a pattern", []string{"git grep -n 'TODO'"}},
			{"git gc", "Clean up and optimize the local repository", nil},
			{"git count-objects -vH", "Show disk usage of object database", nil},
			{"git bundle create repo.bundle --all", "Bundle entire repo into a single file (offline transfer)", nil},
			{"git format-patch HEAD~5", "Generate patch files for the last 5 commits", nil},
			{"git apply <patch>", "Apply a patch file", nil},
			{"git am <patch>", "Apply a patch file preserving commit info", nil},
			{"git notes add -m \"msg\" <commit>", "Attach a note to a commit", nil},
			{"git filter-branch --tree-filter 'rm -f <file>' HEAD", "Remove a file from all commits (use git-filter-repo for large repos)", nil},
		},
	},
}

func (s *Server) GitCheatSheet(_ context.Context, req *pb.GitCheatSheetRequest) (*pb.GitCheatSheetResponse, error) {
	q := strings.ToLower(strings.TrimSpace(req.Query))
	catFilter := strings.ToLower(strings.TrimSpace(req.Category))

	var categories []*pb.GitCmdCategory
	for _, cat := range gitData {
		if catFilter != "" && !strings.Contains(strings.ToLower(cat.name), catFilter) {
			continue
		}

		var cmds []*pb.GitCmd
		for _, cmd := range cat.commands {
			if q != "" {
				if !strings.Contains(strings.ToLower(cmd.command), q) &&
					!strings.Contains(strings.ToLower(cmd.description), q) {
					matched := false
					for _, ex := range cmd.examples {
						if strings.Contains(strings.ToLower(ex), q) {
							matched = true
							break
						}
					}
					if !matched {
						continue
					}
				}
			}
			cmds = append(cmds, &pb.GitCmd{
				Command:     cmd.command,
				Description: cmd.description,
				Examples:    cmd.examples,
			})
		}
		if len(cmds) > 0 {
			categories = append(categories, &pb.GitCmdCategory{
				Name:     cat.name,
				Commands: cmds,
			})
		}
	}
	return &pb.GitCheatSheetResponse{Categories: categories}, nil
}
