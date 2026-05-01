package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	pb "github.com/odinnordico/privutil/proto"
	goexif "github.com/rwcarlsen/goexif/exif"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── SVG Optimizer ────────────────────────────────────────────────────────────

var (
	svgReXmlDecl       = regexp.MustCompile(`(?s)<\?xml[^?]*\?>`)
	svgReDoctype       = regexp.MustCompile(`(?s)<!DOCTYPE[^\[>]*(?:\[[^\]]*\])?>`)
	svgReComment       = regexp.MustCompile(`(?s)<!--.*?-->`)
	svgReMetadata      = regexp.MustCompile(`(?si)<metadata[\s>][\s\S]*?</metadata>`)
	svgReTitle         = regexp.MustCompile(`(?si)<title[\s>][\s\S]*?</title>`)
	svgReDesc          = regexp.MustCompile(`(?si)<desc[\s>][\s\S]*?</desc>`)
	svgReEmptyGroup    = regexp.MustCompile(`(?si)<(g|defs|symbol|marker|clipPath|mask|pattern|linearGradient|radialGradient|filter|switch|a)(\s[^>]*)?>[ \t\r\n]*</(g|defs|symbol|marker|clipPath|mask|pattern|linearGradient|radialGradient|filter|switch|a)>`)
	svgReEmptySelfG    = regexp.MustCompile(`(?si)<(g|defs|symbol)(\s[^>]*)?/>`)
	svgReInterTag      = regexp.MustCompile(`>[ \t\r\n]+<`)
	svgReEmptyStyle    = regexp.MustCompile(` style=""`)
	svgReEmptyClass    = regexp.MustCompile(` class=""`)
	svgReEmptyId       = regexp.MustCompile(` id=""`)
	svgReTrailingSpace = regexp.MustCompile(`[ \t]+\n`)
	svgReMultiNewline  = regexp.MustCompile(`\n{3,}`)
)

type svgOpts struct {
	removeComments     bool
	removeXmlDecl      bool
	removeDoctype      bool
	removeMetadata     bool
	removeTitle        bool
	removeDesc         bool
	removeEmptyGroups  bool
	collapseWhitespace bool
	removeEmptyAttrs   bool
}

func svgOptsFromPreset(preset string) svgOpts {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "aggressive":
		return svgOpts{true, true, true, true, true, true, true, true, true}
	case "minimal":
		return svgOpts{true, false, false, false, false, false, false, false, false}
	default: // "safe" or ""
		return svgOpts{true, true, true, true, false, false, true, true, true}
	}
}

func (s *Server) SvgOptimize(_ context.Context, req *pb.SvgOptimizeRequest) (*pb.SvgOptimizeResponse, error) {
	if strings.TrimSpace(req.Svg) == "" {
		return nil, status.Error(codes.InvalidArgument, "svg is required")
	}

	var opts svgOpts
	if strings.ToLower(req.Preset) == "custom" {
		opts = svgOpts{
			removeComments:     req.RemoveComments,
			removeXmlDecl:      req.RemoveXmlDecl,
			removeDoctype:      req.RemoveDoctype,
			removeMetadata:     req.RemoveMetadata,
			removeTitle:        req.RemoveTitle,
			removeDesc:         req.RemoveDesc,
			removeEmptyGroups:  req.RemoveEmptyGroups,
			collapseWhitespace: req.CollapseWhitespace,
			removeEmptyAttrs:   req.RemoveEmptyAttrs,
		}
	} else {
		opts = svgOptsFromPreset(req.Preset)
	}

	original := req.Svg
	result := original
	var applied []string

	apply := func(name string, fn func(string) string) {
		before := result
		result = fn(result)
		if result != before {
			applied = append(applied, name)
		}
	}

	if opts.removeXmlDecl {
		apply("Remove XML declaration", func(s string) string { return svgReXmlDecl.ReplaceAllString(s, "") })
	}
	if opts.removeDoctype {
		apply("Remove DOCTYPE", func(s string) string { return svgReDoctype.ReplaceAllString(s, "") })
	}
	if opts.removeComments {
		apply("Remove comments", func(s string) string { return svgReComment.ReplaceAllString(s, "") })
	}
	if opts.removeMetadata {
		apply("Remove <metadata>", func(s string) string { return svgReMetadata.ReplaceAllString(s, "") })
	}
	if opts.removeTitle {
		apply("Remove <title>", func(s string) string { return svgReTitle.ReplaceAllString(s, "") })
	}
	if opts.removeDesc {
		apply("Remove <desc>", func(s string) string { return svgReDesc.ReplaceAllString(s, "") })
	}
	if opts.removeEmptyAttrs {
		apply("Remove empty attributes", func(s string) string {
			s = svgReEmptyStyle.ReplaceAllString(s, "")
			s = svgReEmptyClass.ReplaceAllString(s, "")
			s = svgReEmptyId.ReplaceAllString(s, "")
			return s
		})
	}
	if opts.collapseWhitespace {
		apply("Collapse whitespace", func(s string) string {
			s = svgReInterTag.ReplaceAllString(s, "><")
			s = svgReTrailingSpace.ReplaceAllString(s, "\n")
			s = svgReMultiNewline.ReplaceAllString(s, "\n\n")
			return strings.TrimSpace(s)
		})
	}
	if opts.removeEmptyGroups {
		apply("Remove empty groups", func(s string) string {
			// Iterative pass to catch newly-emptied groups after whitespace collapse
			for {
				prev := s
				s = svgReEmptyGroup.ReplaceAllString(s, "")
				s = svgReEmptySelfG.ReplaceAllString(s, "")
				if s == prev {
					break
				}
			}
			return s
		})
	}
	// Always trim
	result = strings.TrimSpace(result)

	origSize := int32(len(original)) // #nosec G115 -- SVG sizes are not realistically >2GiB
	optSize := int32(len(result))   // #nosec G115
	var pct float32
	if origSize > 0 {
		pct = float32(origSize-optSize) / float32(origSize) * 100
	}

	return &pb.SvgOptimizeResponse{
		Result:        result,
		OriginalSize:  origSize,
		OptimizedSize: optSize,
		SavingsPct:    pct,
		Applied:       applied,
	}, nil
}

// ─── Image Metadata / EXIF ────────────────────────────────────────────────────

type exifTagEntry struct {
	tag   goexif.FieldName
	label string
	group string
}

// ordered list of tags we care about
var exifTagList = []exifTagEntry{
	{goexif.Make, "Make", "Camera"},
	{goexif.Model, "Model", "Camera"},
	{goexif.Software, "Software", "Camera"},
	{goexif.Artist, "Artist", "Camera"},
	{goexif.Copyright, "Copyright", "Camera"},
	{goexif.LensMake, "Lens Make", "Camera"},
	{goexif.LensModel, "Lens Model", "Camera"},
	{goexif.PixelXDimension, "Width (EXIF)", "Image"},
	{goexif.PixelYDimension, "Height (EXIF)", "Image"},
	{goexif.Orientation, "Orientation", "Image"},
	{goexif.XResolution, "X Resolution", "Image"},
	{goexif.YResolution, "Y Resolution", "Image"},
	{goexif.ResolutionUnit, "Resolution Unit", "Image"},
	{goexif.ColorSpace, "Color Space", "Image"},
	{goexif.DateTime, "Modified", "DateTime"},
	{goexif.DateTimeOriginal, "Taken", "DateTime"},
	{goexif.DateTimeDigitized, "Digitized", "DateTime"},
	{goexif.ExposureTime, "Exposure Time", "Settings"},
	{goexif.FNumber, "F-Number", "Settings"},
	{goexif.ISOSpeedRatings, "ISO", "Settings"},
	{goexif.FocalLength, "Focal Length", "Settings"},
	{goexif.FocalLengthIn35mmFilm, "Focal Length (35mm eq.)", "Settings"},
	{goexif.Flash, "Flash", "Settings"},
	{goexif.WhiteBalance, "White Balance", "Settings"},
	{goexif.ExposureProgram, "Exposure Program", "Settings"},
	{goexif.MeteringMode, "Metering Mode", "Settings"},
	{goexif.ExposureBiasValue, "Exposure Bias", "Settings"},
	{goexif.MaxApertureValue, "Max Aperture", "Settings"},
	{goexif.SceneCaptureType, "Scene Type", "Settings"},
}

func extractExifFields(x *goexif.Exif) []*pb.ExifField {
	var fields []*pb.ExifField
	for _, entry := range exifTagList {
		tag, err := x.Get(entry.tag)
		if err != nil {
			continue
		}
		val := tag.String()
		// goexif wraps strings in quotes — strip them for display
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		fields = append(fields, &pb.ExifField{Label: entry.label, Value: val, Group: entry.group})
	}
	return fields
}

func dmsStr(deg, min, sec float64, ref string) string {
	return fmt.Sprintf("%d°%d'%.2f\"%s", int(deg), int(min), sec, ref)
}

func detectFormat(data []byte, filename string) string {
	if len(data) < 4 {
		return "unknown"
	}
	switch {
	case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "jpeg"
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47:
		return "png"
	case len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "webp"
	case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
		return "gif"
	case data[0] == 0x42 && data[1] == 0x4D:
		return "bmp"
	}
	// Fallback to extension
	ext := strings.ToLower(strings.TrimPrefix(filename, "."))
	if idx := strings.LastIndex(ext, "."); idx >= 0 {
		ext = ext[idx+1:]
	}
	return ext
}

func parsePNG(data []byte) (width, height int32, fields []*pb.ExifField) {
	if len(data) < 24 {
		return 0, 0, nil
	}
	// IHDR is always the first chunk at offset 8
	// 8 bytes signature + 4 (length) + 4 (type) + data
	if string(data[12:16]) == "IHDR" && len(data) >= 24 {
		width = int32(binary.BigEndian.Uint32(data[16:20]))  // #nosec G115 -- PNG spec limits dimensions to ≤2^31-1
		height = int32(binary.BigEndian.Uint32(data[20:24])) // #nosec G115
		if len(data) >= 25 {
			bitDepth := data[24]
			fields = append(fields, &pb.ExifField{Label: "Bit Depth", Value: strconv.Itoa(int(bitDepth)), Group: "Image"})
		}
		if len(data) >= 26 {
			colorType := data[25]
			ct := map[byte]string{
				0: "Grayscale", 2: "RGB", 3: "Indexed", 4: "Grayscale+Alpha", 6: "RGBA",
			}
			if name, ok := ct[colorType]; ok {
				fields = append(fields, &pb.ExifField{Label: "Color Type", Value: name, Group: "Image"})
			}
		}
	}

	// Walk remaining chunks
	offset := 8
	for offset+12 <= len(data) {
		chunkLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		if offset+8+chunkLen > len(data) {
			break
		}
		chunkType := string(data[offset+4 : offset+8])
		chunkData := data[offset+8 : offset+8+chunkLen]

		switch chunkType {
		case "tEXt":
			parts := bytes.SplitN(chunkData, []byte{0}, 2)
			if len(parts) == 2 {
				fields = append(fields, &pb.ExifField{
					Label: string(parts[0]), Value: string(parts[1]), Group: "Metadata",
				})
			}
		case "iTXt":
			parts := bytes.SplitN(chunkData, []byte{0}, 6)
			if len(parts) >= 6 {
				fields = append(fields, &pb.ExifField{
					Label: string(parts[0]), Value: string(parts[5]), Group: "Metadata",
				})
			}
		case "pHYs":
			if len(chunkData) >= 9 {
				xRes := binary.BigEndian.Uint32(chunkData[0:4])
				yRes := binary.BigEndian.Uint32(chunkData[4:8])
				unit := chunkData[8]
				if unit == 1 {
					xDpi := math.Round(float64(xRes) * 0.0254)
					yDpi := math.Round(float64(yRes) * 0.0254)
					fields = append(fields, &pb.ExifField{
						Label: "DPI", Value: fmt.Sprintf("%.0f × %.0f", xDpi, yDpi), Group: "Image",
					})
				} else if unit == 0 {
					fields = append(fields, &pb.ExifField{
						Label: "Pixel Ratio", Value: fmt.Sprintf("%d × %d (aspect only)", xRes, yRes), Group: "Image",
					})
				}
			}
		case "eXIf":
			// Embedded EXIF in PNG (newer spec)
			if x, err := goexif.Decode(bytes.NewReader(chunkData)); err == nil {
				fields = append(fields, extractExifFields(x)...)
			}
		case "IEND":
			goto done
		}
		offset += 4 + 4 + chunkLen + 4 // length + type + data + CRC
	}
done:
	return width, height, fields
}

func parseWebP(data []byte) (width, height int32, fields []*pb.ExifField) {
	if len(data) < 12 {
		return 0, 0, nil
	}
	offset := 12
	for offset+8 <= len(data) {
		chunkType := string(data[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		if offset+8+chunkSize > len(data) {
			break
		}
		chunkData := data[offset+8 : offset+8+chunkSize]

		switch chunkType {
		case "VP8 ":
			if len(chunkData) >= 10 && chunkData[3] == 0x9D && chunkData[4] == 0x01 && chunkData[5] == 0x2A {
				width = int32(binary.LittleEndian.Uint16(chunkData[6:8])) & 0x3FFF
				height = int32(binary.LittleEndian.Uint16(chunkData[8:10])) & 0x3FFF
				fields = append(fields, &pb.ExifField{Label: "Encoding", Value: "Lossy (VP8)", Group: "Image"})
			}
		case "VP8L":
			if len(chunkData) >= 5 && chunkData[0] == 0x2F {
				bits := uint32(chunkData[1]) | uint32(chunkData[2])<<8 | uint32(chunkData[3])<<16 | uint32(chunkData[4])<<24
				width = int32(bits&0x3FFF) + 1
				height = int32((bits>>14)&0x3FFF) + 1
				fields = append(fields, &pb.ExifField{Label: "Encoding", Value: "Lossless (VP8L)", Group: "Image"})
			}
		case "VP8X":
			if len(chunkData) >= 10 {
				width = int32(uint32(chunkData[4])|uint32(chunkData[5])<<8|uint32(chunkData[6])<<16) + 1  // #nosec G115 -- VP8X uses 24-bit values, max 16777216
				height = int32(uint32(chunkData[7])|uint32(chunkData[8])<<8|uint32(chunkData[9])<<16) + 1 // #nosec G115
				flags := chunkData[0]
				var feats []string
				if flags&0x02 != 0 {
					feats = append(feats, "ICC")
				}
				if flags&0x04 != 0 {
					feats = append(feats, "Alpha")
				}
				if flags&0x08 != 0 {
					feats = append(feats, "EXIF")
				}
				if flags&0x10 != 0 {
					feats = append(feats, "XMP")
				}
				if flags&0x20 != 0 {
					feats = append(feats, "Animation")
				}
				if len(feats) > 0 {
					fields = append(fields, &pb.ExifField{Label: "Features", Value: strings.Join(feats, ", "), Group: "Image"})
				}
			}
		case "EXIF":
			// Possibly TIFF-header-prefixed; try directly first
			r := bytes.NewReader(chunkData)
			if x, err := goexif.Decode(r); err == nil {
				fields = append(fields, extractExifFields(x)...)
			}
		}

		offset += 8 + chunkSize
		if chunkSize%2 != 0 {
			offset++
		}
	}
	return width, height, fields
}

func (s *Server) ExifRead(_ context.Context, req *pb.ExifReadRequest) (*pb.ExifReadResponse, error) {
	if len(req.Data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}
	const maxSize = 10 * 1024 * 1024
	if len(req.Data) > maxSize {
		return nil, status.Errorf(codes.InvalidArgument, "file too large (max 10 MB, got %d bytes)", len(req.Data))
	}

	format := detectFormat(req.Data, req.Filename)
	res := &pb.ExifReadResponse{Format: format}
	var extraFields []*pb.ExifField

	switch format {
	case "jpeg":
		// Primary EXIF decode via goexif
		x, err := func() (ex *goexif.Exif, e error) {
			defer func() {
				if r := recover(); r != nil {
					e = fmt.Errorf("exif panic: %v", r)
				}
			}()
			return goexif.Decode(bytes.NewReader(req.Data))
		}()
		if err == nil {
			// Dimensions from EXIF tags
			if t, err2 := x.Get(goexif.PixelXDimension); err2 == nil {
				if v, err3 := t.Int(0); err3 == nil {
					res.Width = int32(v) // #nosec G115 -- image pixel dimensions are not realistically >2^31
				}
			}
			if t, err2 := x.Get(goexif.PixelYDimension); err2 == nil {
				if v, err3 := t.Int(0); err3 == nil {
					res.Height = int32(v) // #nosec G115
				}
			}
			// GPS
			lat, lng, gpsErr := x.LatLong()
			if gpsErr == nil {
				res.GpsDecimal = fmt.Sprintf("%.6f, %.6f", lat, lng)
				latRef := "N"
				if lat < 0 {
					latRef = "S"
					lat = -lat
				}
				lngRef := "E"
				if lng < 0 {
					lngRef = "W"
					lng = -lng
				}
				latD, latM, latS := decimalToDMS(lat)
				lngD, lngM, lngS := decimalToDMS(lng)
				res.GpsDms = fmt.Sprintf("%s, %s", dmsStr(latD, latM, latS, latRef), dmsStr(lngD, lngM, lngS, lngRef))
				res.MapsUrl = fmt.Sprintf("https://www.google.com/maps?q=%s", res.GpsDecimal)

				// Altitude
				if alt, altErr := x.Get(goexif.GPSAltitude); altErr == nil {
					extraFields = append(extraFields, &pb.ExifField{Label: "GPS Altitude", Value: alt.String() + " m", Group: "GPS"})
				}
				extraFields = append(extraFields, &pb.ExifField{
					Label: "GPS Coordinates (decimal)", Value: res.GpsDecimal, Group: "GPS",
				})
				extraFields = append(extraFields, &pb.ExifField{
					Label: "GPS Coordinates (DMS)", Value: res.GpsDms, Group: "GPS",
				})
			}
			res.Fields = append(res.Fields, extractExifFields(x)...)
		}
		// Fallback: parse JPEG SOF for dimensions if EXIF didn't have them
		if res.Width == 0 || res.Height == 0 {
			w, h := jpegDimensions(req.Data)
			if res.Width == 0 {
				res.Width = w
			}
			if res.Height == 0 {
				res.Height = h
			}
		}

	case "png":
		w, h, fields := parsePNG(req.Data)
		res.Width, res.Height = w, h
		res.Fields = fields

	case "webp":
		w, h, fields := parseWebP(req.Data)
		res.Width, res.Height = w, h
		res.Fields = fields

	default:
		res.Error = fmt.Sprintf("unsupported format: %s (supports JPEG, PNG, WebP)", format)
		return res, nil
	}

	// Append dimension fields if we got them
	if res.Width > 0 && res.Height > 0 {
		dimField := &pb.ExifField{Label: "Dimensions", Value: fmt.Sprintf("%d × %d px", res.Width, res.Height), Group: "Image"}
		// Prepend dimensions
		res.Fields = append([]*pb.ExifField{dimField}, res.Fields...)
	}
	res.Fields = append(res.Fields, extraFields...)

	if len(res.Fields) == 0 && res.Error == "" {
		res.Error = "no metadata found in this file"
	}
	return res, nil
}

func decimalToDMS(decimal float64) (deg, min, sec float64) {
	deg = math.Floor(decimal)
	minFull := (decimal - deg) * 60
	min = math.Floor(minFull)
	sec = (minFull - min) * 60
	return deg, min, sec
}

func jpegDimensions(data []byte) (width, height int32) {
	i := 2 // Skip SOI marker FF D8
	for i+4 <= len(data) {
		if data[i] != 0xFF {
			break
		}
		marker := data[i+1]
		segLen := int(data[i+2])<<8 | int(data[i+3])
		// SOF markers: C0-C3, C5-C7, C9-CB, CD-CF
		if (marker >= 0xC0 && marker <= 0xCF) && marker != 0xC4 && marker != 0xC8 && marker != 0xCC {
			if i+9 <= len(data) {
				height = int32(data[i+5])<<8 | int32(data[i+6])
				width = int32(data[i+7])<<8 | int32(data[i+8])
				return width, height
			}
		}
		i += 2 + segLen
	}
	return 0, 0
}

// ─── File ↔ Base64 ────────────────────────────────────────────────────────────

var mimeToExt = map[string]string{
	"image/jpeg":       "jpg",
	"image/png":        "png",
	"image/gif":        "gif",
	"image/webp":       "webp",
	"image/svg+xml":    "svg",
	"image/bmp":        "bmp",
	"image/tiff":       "tif",
	"image/x-icon":     "ico",
	"audio/mpeg":       "mp3",
	"audio/ogg":        "ogg",
	"audio/wav":        "wav",
	"audio/flac":       "flac",
	"video/mp4":        "mp4",
	"video/webm":       "webm",
	"application/pdf":  "pdf",
	"application/zip":  "zip",
	"application/gzip": "gz",
	"application/json": "json",
	"text/plain":       "txt",
	"text/html":        "html",
	"text/css":         "css",
	"text/csv":         "csv",
	"font/ttf":         "ttf",
	"font/woff":        "woff",
	"font/woff2":       "woff2",
}

func (s *Server) FileToBase64(_ context.Context, req *pb.FileToBase64Request) (*pb.FileToBase64Response, error) {
	if len(req.Data) == 0 {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}
	const maxSize = 10 * 1024 * 1024
	if len(req.Data) > maxSize {
		return nil, status.Errorf(codes.InvalidArgument, "file too large (max 10 MB)")
	}

	mimeType := http.DetectContentType(req.Data)
	// DetectContentType returns things like "image/jpeg; charset=..." — strip params
	if idx := strings.Index(mimeType, ";"); idx >= 0 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}

	encoded := base64.StdEncoding.EncodeToString(req.Data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	return &pb.FileToBase64Response{
		Encoded:  encoded,
		DataUri:  dataURI,
		MimeType: mimeType,
		Size:     int32(len(req.Data)), // #nosec G115 -- file sizes are not realistically >2GiB in this context
	}, nil
}

func (s *Server) Base64ToFile(_ context.Context, req *pb.Base64ToFileRequest) (*pb.Base64ToFileResponse, error) {
	if strings.TrimSpace(req.Encoded) == "" {
		return nil, status.Error(codes.InvalidArgument, "encoded is required")
	}

	encoded := strings.TrimSpace(req.Encoded)
	mimeHint := ""

	// Handle data URI: "data:mime/type;base64,<data>"
	if strings.HasPrefix(encoded, "data:") {
		comma := strings.Index(encoded, ",")
		if comma < 0 {
			return &pb.Base64ToFileResponse{Error: "invalid data URI: missing comma"}, nil
		}
		header := encoded[5:comma] // strip "data:"
		encoded = encoded[comma+1:]
		parts := strings.SplitN(header, ";", 2)
		if len(parts) >= 1 {
			mimeHint = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 && !strings.EqualFold(strings.TrimSpace(parts[1]), "base64") {
			return &pb.Base64ToFileResponse{Error: "only base64 data URIs are supported"}, nil
		}
	}

	// Normalize: remove whitespace/newlines that might be in multiline base64
	encoded = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, encoded)

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Try URL-safe variant
		decoded, err = base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			// Try without padding
			decoded, err = base64.RawStdEncoding.DecodeString(encoded)
			if err != nil {
				return &pb.Base64ToFileResponse{Error: "invalid base64: " + err.Error()}, nil
			}
		}
	}

	mimeType := http.DetectContentType(decoded)
	if idx := strings.Index(mimeType, ";"); idx >= 0 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}
	// If DetectContentType returns generic "application/octet-stream", use the hint
	if mimeType == "application/octet-stream" && mimeHint != "" {
		mimeType = mimeHint
	}

	// Build suggested filename
	filename := req.Filename
	if filename == "" {
		ext := mimeToExt[mimeType]
		if ext == "" {
			ext = "bin"
		}
		filename = "decoded." + ext
	}

	return &pb.Base64ToFileResponse{
		Data:     decoded,
		MimeType: mimeType,
		Filename: filename,
		Size:     int32(len(decoded)), // #nosec G115 -- file sizes are not realistically >2GiB in this context
	}, nil
}
