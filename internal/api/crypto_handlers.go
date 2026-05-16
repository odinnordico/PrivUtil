package api

import (
	"context"
	"crypto/hmac"
	"crypto/md5" // #nosec G501 -- user-selectable HMAC; weak hash is caller's choice
	"crypto/rand"
	"crypto/sha1" // #nosec G505 -- user-selectable HMAC; weak hash is caller's choice
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"math"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	pb "github.com/odinnordico/privutil/proto"
	ulidpkg "github.com/oklog/ulid/v2"
)

// ─── HMAC ─────────────────────────────────────────────────────────────────────

func (s *Server) HmacGenerate(_ context.Context, req *pb.HmacRequest) (*pb.HmacResponse, error) {
	var h hash.Hash
	switch strings.ToLower(req.Algo) {
	case "sha512":
		h = hmac.New(sha512.New, []byte(req.Secret))
	case "sha1":
		h = hmac.New(sha1.New, []byte(req.Secret)) //nolint:gosec
	case "md5":
		h = hmac.New(md5.New, []byte(req.Secret)) //nolint:gosec
	default: // "sha256" or empty
		h = hmac.New(sha256.New, []byte(req.Secret))
	}
	h.Write([]byte(req.Message))
	mac := h.Sum(nil)
	return &pb.HmacResponse{
		Hex:    hex.EncodeToString(mac),
		Base64: base64.StdEncoding.EncodeToString(mac),
	}, nil
}

// ─── OTP (TOTP / HOTP) ────────────────────────────────────────────────────────

func otpHash(algo string) func() hash.Hash {
	switch strings.ToLower(algo) {
	case "sha256":
		return sha256.New
	case "sha512":
		return sha512.New
	default:
		return sha1.New //nolint:gosec
	}
}

func hotpCode(secret []byte, counter uint64, digits int, hashFn func() hash.Hash) (string, error) {
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, counter)
	h := hmac.New(hashFn, secret)
	h.Write(msg)
	mac := h.Sum(nil)
	offset := mac[len(mac)-1] & 0x0f
	code := binary.BigEndian.Uint32(mac[offset:offset+4]) & 0x7fffffff
	code %= uint32(math.Pow10(digits))
	return fmt.Sprintf("%0*d", digits, code), nil
}

func decodeOTPSecret(secret string) ([]byte, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	// Pad to multiple of 8
	if pad := len(secret) % 8; pad != 0 {
		secret += strings.Repeat("=", 8-pad)
	}
	return base32.StdEncoding.DecodeString(secret)
}

func (s *Server) OtpGenerate(_ context.Context, req *pb.OtpRequest) (*pb.OtpResponse, error) {
	digits := int(req.Digits)
	if digits != 8 {
		digits = 6
	}
	period := int64(req.Period)
	if period <= 0 {
		period = 30
	}
	hashFn := otpHash(req.Algo)

	var secretBytes []byte
	var secretStr string

	if req.GenerateSecret {
		raw := make([]byte, 20)
		if _, err := rand.Read(raw); err != nil {
			return &pb.OtpResponse{Error: "failed to generate secret"}, nil
		}
		secretStr = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
		secretBytes = raw
	} else {
		secretStr = req.Secret
		var err error
		secretBytes, err = decodeOTPSecret(req.Secret)
		if err != nil {
			return &pb.OtpResponse{Error: "invalid base32 secret: " + err.Error()}, nil
		}
	}

	now := time.Now()
	var code string
	var err error
	var validUntil, timeRemaining int64

	if strings.ToLower(req.Type) == "hotp" {
		if req.Counter < 0 {
			return &pb.OtpResponse{Error: "counter must be non-negative"}, nil
		}
		code, err = hotpCode(secretBytes, uint64(req.Counter), digits, hashFn)
		if err != nil {
			return &pb.OtpResponse{Error: err.Error()}, nil
		}
	} else {
		// TOTP
		counter := now.Unix() / period
		code, err = hotpCode(secretBytes, uint64(counter), digits, hashFn) // #nosec G115 -- counter = Unix()/period, always non-negative
		if err != nil {
			return &pb.OtpResponse{Error: err.Error()}, nil
		}
		periodEnd := (counter + 1) * period
		validUntil = periodEnd
		timeRemaining = periodEnd - now.Unix()
	}

	issuer := "privutil"
	algoLabel := strings.ToUpper(req.Algo)
	if algoLabel == "" {
		algoLabel = "SHA1"
	}
	uri := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s&algorithm=%s&digits=%d&period=%d",
		issuer, secretStr, issuer, algoLabel, digits, period)

	return &pb.OtpResponse{
		Code:          code,
		Secret:        secretStr,
		ValidUntil:    validUntil,
		TimeRemaining: timeRemaining,
		Uri:           uri,
	}, nil
}

func (s *Server) OtpValidate(_ context.Context, req *pb.OtpValidateRequest) (*pb.OtpValidateResponse, error) {
	secretBytes, err := decodeOTPSecret(req.Secret)
	if err != nil {
		return &pb.OtpValidateResponse{Error: "invalid base32 secret: " + err.Error()}, nil
	}

	period := int64(req.Period)
	if period <= 0 {
		period = 30
	}
	window := int(req.Window)
	if window <= 0 {
		window = 1
	}
	hashFn := otpHash(req.Algo)

	counter := time.Now().Unix() / period
	for delta := int64(-window); delta <= int64(window); delta++ {
		code, err := hotpCode(secretBytes, uint64(counter+delta), 6, hashFn) // #nosec G115 -- counter is large Unix time; delta is ±window (default 1)
		if err != nil {
			continue
		}
		if code == strings.TrimSpace(req.Code) {
			return &pb.OtpValidateResponse{Valid: true}, nil
		}
	}
	return &pb.OtpValidateResponse{Valid: false}, nil
}

// ─── ULID ─────────────────────────────────────────────────────────────────────

const maxUlidCount = 100

func (s *Server) UlidGenerate(_ context.Context, req *pb.UlidRequest) (*pb.UlidResponse, error) {
	count := int(req.Count)
	if count <= 0 {
		count = 1
	}
	if count > maxUlidCount {
		count = maxUlidCount
	}

	ulids := make([]string, count)
	entropy := ulidpkg.Monotonic(rand.Reader, 0)

	for i := range ulids {
		id, err := ulidpkg.New(ulidpkg.Timestamp(time.Now()), entropy)
		if err != nil {
			return &pb.UlidResponse{Error: "generation failed: " + err.Error()}, nil
		}
		ulids[i] = id.String()
	}
	return &pb.UlidResponse{Ulids: ulids}, nil
}

// ─── Caesar / ROT13 ───────────────────────────────────────────────────────────

func (s *Server) CaesarCipher(_ context.Context, req *pb.CaesarRequest) (*pb.CaesarResponse, error) {
	if req.Text == "" {
		return &pb.CaesarResponse{Error: "input is required"}, nil
	}
	shift := int(req.Shift) % 26
	if req.Action == "decode" {
		shift = -shift
	}
	shift = ((shift % 26) + 26) % 26

	var sb strings.Builder
	for _, r := range req.Text {
		switch {
		case r >= 'A' && r <= 'Z':
			sb.WriteRune('A' + (r-'A'+rune(shift))%26)
		case r >= 'a' && r <= 'z':
			sb.WriteRune('a' + (r-'a'+rune(shift))%26)
		default:
			sb.WriteRune(r)
		}
	}
	return &pb.CaesarResponse{Result: sb.String()}, nil
}

// ─── Text encoder (binary / hex / octal / decimal) ───────────────────────────

func (s *Server) TextEncode(_ context.Context, req *pb.TextEncodeRequest) (*pb.TextEncodeResponse, error) {
	if req.Text == "" {
		return &pb.TextEncodeResponse{Error: "input is required"}, nil
	}

	format := strings.ToLower(req.Format)
	action := strings.ToLower(req.Action)

	if action == "decode" {
		result, err := textDecode(req.Text, format)
		if err != nil {
			return &pb.TextEncodeResponse{Error: err.Error()}, nil
		}
		return &pb.TextEncodeResponse{Result: result}, nil
	}

	// encode
	result, err := textEncodeBytes(req.Text, format)
	if err != nil {
		return &pb.TextEncodeResponse{Error: err.Error()}, nil
	}
	return &pb.TextEncodeResponse{Result: result}, nil
}

func textEncodeBytes(text, format string) (string, error) {
	if !utf8.ValidString(text) {
		return "", fmt.Errorf("input is not valid UTF-8")
	}
	parts := make([]string, 0, len(text))
	for _, b := range []byte(text) {
		switch format {
		case "binary":
			parts = append(parts, fmt.Sprintf("%08b", b))
		case "octal":
			parts = append(parts, fmt.Sprintf("%03o", b))
		case "decimal":
			parts = append(parts, fmt.Sprintf("%d", b))
		default: // hex
			parts = append(parts, fmt.Sprintf("%02x", b))
		}
	}
	return strings.Join(parts, " "), nil
}

func textDecode(encoded, format string) (string, error) {
	tokens := strings.Fields(encoded)
	if len(tokens) == 0 {
		return "", fmt.Errorf("no input tokens")
	}

	buf := make([]byte, len(tokens))
	for i, tok := range tokens {
		var base int
		switch format {
		case "binary":
			base = 2
		case "octal":
			base = 8
		case "decimal":
			base = 10
		default:
			base = 16
		}
		n := new(big.Int)
		if _, ok := n.SetString(tok, base); !ok {
			return "", fmt.Errorf("invalid token %q for format %s", tok, format)
		}
		if n.IsInt64() && n.Int64() >= 0 && n.Int64() <= 255 {
			buf[i] = byte(n.Int64()) // #nosec G115 -- validated 0..255 above
		} else {
			return "", fmt.Errorf("value %q out of byte range", tok)
		}
	}
	return string(buf), nil
}

// ─── Morse code ───────────────────────────────────────────────────────────────

var morseEnc = map[rune]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".", 'F': "..-.", 'G': "--.",
	'H': "....", 'I': "..", 'J': ".---", 'K': "-.-", 'L': ".-..", 'M': "--", 'N': "-.",
	'O': "---", 'P': ".--.", 'Q': "--.-", 'R': ".-.", 'S': "...", 'T': "-", 'U': "..-",
	'V': "...-", 'W': ".--", 'X': "-..-", 'Y': "-.--", 'Z': "--..",
	'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-", '5': ".....",
	'6': "-....", '7': "--...", '8': "---..", '9': "----.",
	'.': ".-.-.-", ',': "--..--", '?': "..--..", '\'': ".----.", '!': "-.-.--",
	'/': "-..-.", '(': "-.--.", ')': "-.--.-", '&': ".-...", ':': "---...", ';': "-.-.-.",
	'=': "-...-", '+': ".-.-.", '-': "-....-", '_': "..--.-", '"': ".-..-.", '$': "...-..-",
	'@': ".--.-.",
}

var morseDec map[string]rune

func init() {
	morseDec = make(map[string]rune, len(morseEnc))
	for k, v := range morseEnc {
		morseDec[v] = k
	}
}

func (s *Server) MorseCode(_ context.Context, req *pb.MorseRequest) (*pb.MorseResponse, error) {
	if req.Text == "" {
		return &pb.MorseResponse{Error: "input is required"}, nil
	}

	if req.Action == "decode" {
		result, err := morseToText(req.Text)
		if err != nil {
			return &pb.MorseResponse{Error: err.Error()}, nil
		}
		return &pb.MorseResponse{Result: result}, nil
	}

	result, err := textToMorse(req.Text)
	if err != nil {
		return &pb.MorseResponse{Error: err.Error()}, nil
	}
	return &pb.MorseResponse{Result: result}, nil
}

func textToMorse(text string) (string, error) {
	var parts []string
	for _, r := range strings.ToUpper(text) {
		if r == ' ' {
			parts = append(parts, "/")
			continue
		}
		code, ok := morseEnc[r]
		if !ok {
			return "", fmt.Errorf("character %q has no Morse code", r)
		}
		parts = append(parts, code)
	}
	return strings.Join(parts, " "), nil
}

func morseToText(morse string) (string, error) {
	var sb strings.Builder
	words := strings.Split(morse, " / ")
	for wi, word := range words {
		if wi > 0 {
			sb.WriteRune(' ')
		}
		codes := strings.FieldsSeq(word)
		for code := range codes {
			r, ok := morseDec[code]
			if !ok {
				return "", fmt.Errorf("unknown Morse code %q", code)
			}
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

// ─── Basic Auth ───────────────────────────────────────────────────────────────

func (s *Server) BasicAuthGenerate(_ context.Context, req *pb.BasicAuthRequest) (*pb.BasicAuthResponse, error) {
	if req.Username == "" {
		return &pb.BasicAuthResponse{Error: "username is required"}, nil
	}
	credentials := req.Username + ":" + req.Password
	token := base64.StdEncoding.EncodeToString([]byte(credentials))
	return &pb.BasicAuthResponse{
		Header:  "Basic " + token,
		Token:   token,
		Decoded: credentials,
	}, nil
}
