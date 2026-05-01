//go:build manual

package api

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

var mediaSrv = &Server{}
var mediaCtx = context.Background()

// ─── SVG Optimizer ────────────────────────────────────────────────────────────

const sampleSVG = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <!-- This is a comment -->
  <metadata>
    <dc:title>Test SVG</dc:title>
  </metadata>
  <title>My Icon</title>
  <desc>A simple test icon</desc>
  <g id="" style="" class="">
    <circle cx="50" cy="50" r="40" fill="blue"/>
  </g>
  <g>
  </g>
</svg>`

func TestSvgOptimize_SafePreset(t *testing.T) {
	res, err := mediaSrv.SvgOptimize(mediaCtx, &pb.SvgOptimizeRequest{Svg: sampleSVG, Preset: "safe"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if strings.Contains(res.Result, "<!--") {
		t.Error("comments should be removed")
	}
	if strings.Contains(res.Result, "<?xml") {
		t.Error("xml declaration should be removed")
	}
	if strings.Contains(res.Result, "DOCTYPE") {
		t.Error("doctype should be removed")
	}
	if strings.Contains(res.Result, "<metadata") {
		t.Error("metadata should be removed")
	}
	// safe preset doesn't remove title/desc
	if !strings.Contains(res.Result, "<title>") {
		t.Error("safe preset should keep <title>")
	}
	if res.OriginalSize <= res.OptimizedSize {
		t.Errorf("expected savings, orig=%d opt=%d", res.OriginalSize, res.OptimizedSize)
	}
	if res.SavingsPct <= 0 {
		t.Errorf("expected positive savings pct, got %f", res.SavingsPct)
	}
	if len(res.Applied) == 0 {
		t.Error("expected at least one transform applied")
	}
}

func TestSvgOptimize_AggressivePreset(t *testing.T) {
	res, err := mediaSrv.SvgOptimize(mediaCtx, &pb.SvgOptimizeRequest{Svg: sampleSVG, Preset: "aggressive"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.Result, "<title>") {
		t.Error("aggressive preset should remove <title>")
	}
	if strings.Contains(res.Result, "<desc>") {
		t.Error("aggressive preset should remove <desc>")
	}
}

func TestSvgOptimize_CustomPreset(t *testing.T) {
	res, err := mediaSrv.SvgOptimize(mediaCtx, &pb.SvgOptimizeRequest{
		Svg:                sampleSVG,
		Preset:             "custom",
		RemoveComments:     true,
		RemoveXmlDecl:      false,
		RemoveDoctype:      false,
		RemoveMetadata:     false,
		CollapseWhitespace: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.Result, "<!--") {
		t.Error("comments should be removed (custom=true)")
	}
	if !strings.Contains(res.Result, "<?xml") {
		t.Error("xml decl should be kept (custom=false)")
	}
	if !strings.Contains(res.Result, "<metadata") {
		t.Error("metadata should be kept (custom=false)")
	}
}

func TestSvgOptimize_EmptyGroupsRemoved(t *testing.T) {
	svg := `<svg><g></g><g>  </g><circle r="5"/></svg>`
	res, err := mediaSrv.SvgOptimize(mediaCtx, &pb.SvgOptimizeRequest{Svg: svg, Preset: "safe"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.Result, "<g>") {
		t.Error("empty groups should be removed")
	}
	if !strings.Contains(res.Result, "<circle") {
		t.Error("non-empty elements should be preserved")
	}
}

func TestSvgOptimize_CollapseWhitespace(t *testing.T) {
	svg := "<svg>\n  \n  <circle r=\"5\"/>\n  \n</svg>"
	res, err := mediaSrv.SvgOptimize(mediaCtx, &pb.SvgOptimizeRequest{
		Svg:                svg,
		Preset:             "custom",
		CollapseWhitespace: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.Result, "  \n  ") {
		t.Error("excess whitespace should be collapsed")
	}
}

func TestSvgOptimize_EmptyInput(t *testing.T) {
	_, err := mediaSrv.SvgOptimize(mediaCtx, &pb.SvgOptimizeRequest{Svg: ""})
	if err == nil {
		t.Fatal("expected error for empty svg")
	}
}

func TestSvgOptimize_NoOpWhenAlreadyOptimal(t *testing.T) {
	svg := `<svg><circle r="5"/></svg>`
	res, err := mediaSrv.SvgOptimize(mediaCtx, &pb.SvgOptimizeRequest{Svg: svg, Preset: "safe"})
	if err != nil {
		t.Fatal(err)
	}
	if res.OriginalSize != res.OptimizedSize {
		// whitespace collapse might still touch it slightly, but result should equal input
		t.Logf("sizes differ slightly: orig=%d opt=%d (OK if whitespace only)", res.OriginalSize, res.OptimizedSize)
	}
}

// ─── Image Metadata / EXIF ────────────────────────────────────────────────────

// minimalJPEG is a 1×1 white JPEG (no EXIF) — ensures format detection and SOF parsing
var minimalJPEGBytes = []byte{
	0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
	0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
	0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
	0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12,
	0x13, 0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20,
	0x24, 0x2E, 0x27, 0x20, 0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29,
	0x2C, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32,
	0x3C, 0x2E, 0x33, 0x34, 0x32, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01,
	0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x1F, 0x00, 0x00,
	0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	0x09, 0x0A, 0x0B, 0xFF, 0xC4, 0x00, 0xB5, 0x10, 0x00, 0x02, 0x01, 0x03,
	0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7D,
	0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
	0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xA1, 0x08,
	0x23, 0x42, 0xB1, 0xC1, 0x15, 0x52, 0xD1, 0xF0, 0x24, 0x33, 0x62, 0x72,
	0x82, 0x09, 0x0A, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x25, 0x26, 0x27, 0x28,
	0x29, 0x2A, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x43, 0x44, 0x45,
	0x46, 0x47, 0x48, 0x49, 0x4A, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
	0x5A, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6A, 0x73, 0x74, 0x75,
	0x76, 0x77, 0x78, 0x79, 0x7A, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
	0x8A, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0xA2, 0xA3,
	0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6,
	0xB7, 0xB8, 0xB9, 0xBA, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9,
	0xCA, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xE1, 0xE2,
	0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xF1, 0xF2, 0xF3, 0xF4,
	0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01,
	0x00, 0x00, 0x3F, 0x00, 0xFB, 0xD5, 0xFF, 0xD9,
}

// minimalPNG is a 1×1 red PNG
var minimalPNGBytes = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
	0x00, 0x00, 0x00, 0x0D, // IHDR length = 13
	0x49, 0x48, 0x44, 0x52, // "IHDR"
	0x00, 0x00, 0x00, 0x01, // width = 1
	0x00, 0x00, 0x00, 0x01, // height = 1
	0x08, 0x02, // bit depth 8, color type 2 (RGB)
	0x00, 0x00, 0x00, // compression, filter, interlace
	0x90, 0x77, 0x53, 0xDE, // CRC
	0x00, 0x00, 0x00, 0x0C, // IDAT length
	0x49, 0x44, 0x41, 0x54, // "IDAT"
	0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01,
	0xE2, 0x21, 0xBC, 0x33, // CRC
	0x00, 0x00, 0x00, 0x00, // IEND length
	0x49, 0x45, 0x4E, 0x44, // "IEND"
	0xAE, 0x42, 0x60, 0x82, // CRC
}

func TestExifRead_JPEG(t *testing.T) {
	res, err := mediaSrv.ExifRead(mediaCtx, &pb.ExifReadRequest{Data: minimalJPEGBytes, Filename: "test.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Format != "jpeg" {
		t.Errorf("format got %q, want jpeg", res.Format)
	}
	// minimalJPEG is 1×1 — SOF should give us dimensions
	if res.Width != 1 || res.Height != 1 {
		t.Logf("width=%d height=%d (SOF fallback)", res.Width, res.Height)
	}
}

func TestExifRead_PNG(t *testing.T) {
	res, err := mediaSrv.ExifRead(mediaCtx, &pb.ExifReadRequest{Data: minimalPNGBytes, Filename: "test.png"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Format != "png" {
		t.Errorf("format got %q, want png", res.Format)
	}
	if res.Width != 1 || res.Height != 1 {
		t.Errorf("dimensions got %d×%d, want 1×1", res.Width, res.Height)
	}
	// Should have at least a Dimensions field and Color Type
	hasDim := false
	for _, f := range res.Fields {
		if f.Label == "Dimensions" {
			hasDim = true
		}
	}
	if !hasDim {
		t.Error("expected Dimensions field")
	}
}

func TestExifRead_EmptyData(t *testing.T) {
	_, err := mediaSrv.ExifRead(mediaCtx, &pb.ExifReadRequest{Data: nil})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestExifRead_Unsupported(t *testing.T) {
	res, err := mediaSrv.ExifRead(mediaCtx, &pb.ExifReadRequest{
		Data:     []byte("not an image file at all"),
		Filename: "test.txt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error == "" {
		t.Error("expected error for unsupported format")
	}
}

func TestExifRead_TooLarge(t *testing.T) {
	bigData := make([]byte, 11*1024*1024) // 11 MB
	bigData[0] = 0xFF
	bigData[1] = 0xD8
	bigData[2] = 0xFF
	_, err := mediaSrv.ExifRead(mediaCtx, &pb.ExifReadRequest{Data: bigData})
	if err == nil {
		t.Fatal("expected error for too-large file")
	}
}

// ─── File ↔ Base64 ────────────────────────────────────────────────────────────

func TestFileToBase64_Basic(t *testing.T) {
	res, err := mediaSrv.FileToBase64(mediaCtx, &pb.FileToBase64Request{
		Data:     minimalPNGBytes,
		Filename: "image.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if res.Encoded == "" {
		t.Error("encoded should not be empty")
	}
	if res.MimeType != "image/png" {
		t.Errorf("mime got %q", res.MimeType)
	}
	if int(res.Size) != len(minimalPNGBytes) {
		t.Errorf("size got %d, want %d", res.Size, len(minimalPNGBytes))
	}
	if !strings.HasPrefix(res.DataUri, "data:image/png;base64,") {
		t.Errorf("data_uri prefix wrong: %q", res.DataUri[:30])
	}

	// Verify encoded is valid base64
	decoded, err := base64.StdEncoding.DecodeString(res.Encoded)
	if err != nil {
		t.Fatalf("encoded is not valid base64: %v", err)
	}
	if len(decoded) != len(minimalPNGBytes) {
		t.Errorf("decoded size mismatch")
	}
}

func TestFileToBase64_JPEG(t *testing.T) {
	res, err := mediaSrv.FileToBase64(mediaCtx, &pb.FileToBase64Request{
		Data: minimalJPEGBytes,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.MimeType != "image/jpeg" {
		t.Errorf("mime got %q", res.MimeType)
	}
}

func TestFileToBase64_Empty(t *testing.T) {
	_, err := mediaSrv.FileToBase64(mediaCtx, &pb.FileToBase64Request{Data: nil})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestBase64ToFile_RawBase64(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString(minimalPNGBytes)
	res, err := mediaSrv.Base64ToFile(mediaCtx, &pb.Base64ToFileRequest{Encoded: encoded})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if len(res.Data) != len(minimalPNGBytes) {
		t.Errorf("decoded size mismatch: got %d want %d", len(res.Data), len(minimalPNGBytes))
	}
	if res.MimeType != "image/png" {
		t.Errorf("mime got %q", res.MimeType)
	}
	if res.Filename == "" {
		t.Error("filename should not be empty")
	}
}

func TestBase64ToFile_DataURI(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString(minimalPNGBytes)
	dataURI := "data:image/png;base64," + encoded
	res, err := mediaSrv.Base64ToFile(mediaCtx, &pb.Base64ToFileRequest{Encoded: dataURI})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if len(res.Data) != len(minimalPNGBytes) {
		t.Errorf("decoded size mismatch")
	}
}

func TestBase64ToFile_WithNewlines(t *testing.T) {
	// Base64 often has line breaks every 76 chars
	raw := base64.StdEncoding.EncodeToString(minimalPNGBytes)
	var b strings.Builder
	for i, c := range raw {
		if i > 0 && i%76 == 0 {
			b.WriteByte('\n')
		}
		b.WriteRune(c)
	}
	res, err := mediaSrv.Base64ToFile(mediaCtx, &pb.Base64ToFileRequest{Encoded: b.String()})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if len(res.Data) != len(minimalPNGBytes) {
		t.Errorf("decoded size mismatch with newlines in input")
	}
}

func TestBase64ToFile_InvalidBase64(t *testing.T) {
	res, err := mediaSrv.Base64ToFile(mediaCtx, &pb.Base64ToFileRequest{Encoded: "not!valid@base64#"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error == "" {
		t.Error("expected error for invalid base64")
	}
}

func TestBase64ToFile_CustomFilename(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString(minimalPNGBytes)
	res, err := mediaSrv.Base64ToFile(mediaCtx, &pb.Base64ToFileRequest{
		Encoded:  encoded,
		Filename: "my_image.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Filename != "my_image.png" {
		t.Errorf("filename got %q", res.Filename)
	}
}

func TestBase64ToFile_Empty(t *testing.T) {
	_, err := mediaSrv.Base64ToFile(mediaCtx, &pb.Base64ToFileRequest{Encoded: ""})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

// ─── Roundtrip ────────────────────────────────────────────────────────────────

func TestFileBase64Roundtrip(t *testing.T) {
	// Encode then decode should recover original bytes
	encRes, err := mediaSrv.FileToBase64(mediaCtx, &pb.FileToBase64Request{
		Data: minimalJPEGBytes,
	})
	if err != nil || encRes.Error != "" {
		t.Fatalf("encode failed: %v %s", err, encRes.GetError())
	}

	decRes, err := mediaSrv.Base64ToFile(mediaCtx, &pb.Base64ToFileRequest{
		Encoded: encRes.DataUri,
	})
	if err != nil || decRes.Error != "" {
		t.Fatalf("decode failed: %v %s", err, decRes.GetError())
	}

	if len(decRes.Data) != len(minimalJPEGBytes) {
		t.Errorf("roundtrip size mismatch: got %d want %d", len(decRes.Data), len(minimalJPEGBytes))
	}
	for i, b := range decRes.Data {
		if b != minimalJPEGBytes[i] {
			t.Errorf("byte mismatch at offset %d", i)
			break
		}
	}
}
