package api

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"

	pb "github.com/odinnordico/privutil/proto"
)

func (s *Server) GenerateUuid(ctx context.Context, req *pb.UuidRequest) (*pb.UuidResponse, error) {
	var uuids []string
	count := req.Count
	if count <= 0 {
		count = 1
	}
	if count > 100 {
		count = 100
	}

	// Get the namespace UUID for name-based UUIDs (v3, v5, v8)
	namespace := getNamespace(req.Namespace)

	for i := 0; i < int(count); i++ {
		var u uuid.UUID
		var err error

		switch req.Version {
		case "v1":
			// Version 1: Time-based UUID
			u, err = uuid.NewUUID()
		case "v2":
			// Version 2: DCE Security UUID (uses NewDCEPerson for simplicity)
			u, err = uuid.NewDCEPerson()
		case "v3":
			// Version 3: Name-based UUID using MD5 hashing
			// Using specified namespace with unique data per iteration
			data := []byte(fmt.Sprintf("privutil.uuid.v3.%d.%d", i, uuid.New().ID()))
			u = uuid.NewMD5(namespace, data)
		case "v5":
			// Version 5: Name-based UUID using SHA1 hashing
			// Using specified namespace with unique data per iteration
			data := []byte(fmt.Sprintf("privutil.uuid.v5.%d.%d", i, uuid.New().ID()))
			u = uuid.NewSHA1(namespace, data)
		case "v6":
			// Version 6: Time-ordered UUID
			u, err = uuid.NewV6()
		case "v7":
			// Version 7: Unix Epoch time-based UUID
			u, err = uuid.NewV7()
		case "v8":
			// Version 8: Custom UUID (using SHA256 hash with custom data for uniqueness)
			// V8 is vendor-specific, so we create unique UUIDs using NewHash
			h := sha256.New()
			data := []byte(fmt.Sprintf("privutil.uuid.v8.%d", i))
			u = uuid.NewHash(h, namespace, data, 8)
		default:
			// Default to Version 4: Random UUID
			u, err = uuid.NewRandom()
		}

		if err != nil {
			return nil, err
		}

		str := u.String()
		if !req.Hyphen {
			str = strings.ReplaceAll(str, "-", "")
		}
		if req.Uppercase {
			str = strings.ToUpper(str)
		}
		uuids = append(uuids, str)
	}

	return &pb.UuidResponse{Uuids: uuids}, nil
}

// getNamespace returns the UUID namespace based on the provided string.
// Defaults to NameSpaceDNS if empty or invalid.
func getNamespace(ns string) uuid.UUID {
	switch strings.ToLower(ns) {
	case "url":
		return uuid.NameSpaceURL
	case "oid":
		return uuid.NameSpaceOID
	case "x500":
		return uuid.NameSpaceX500
	default:
		// Default to DNS namespace
		return uuid.NameSpaceDNS
	}
}

func (s *Server) GenerateLorem(ctx context.Context, req *pb.LoremRequest) (*pb.LoremResponse, error) {
	var text string
	count := int(req.Count)
	if count <= 0 {
		count = 1
	}

	switch req.Type {
	case "word":
		var words []string
		for i := 0; i < count; i++ {
			words = append(words, gofakeit.Word())
		}
		text = strings.Join(words, " ")
	case "sentence":
		var sentences []string
		for i := 0; i < count; i++ {
			sentences = append(sentences, gofakeit.Sentence(10))
		}
		text = strings.Join(sentences, " ")
	default:
		var paragraphs []string
		for i := 0; i < count; i++ {
			paragraphs = append(paragraphs, gofakeit.Paragraph(3, 5, 10, "\n"))
		}
		text = strings.Join(paragraphs, "\n\n")
	}

	return &pb.LoremResponse{Text: text}, nil
}

func (s *Server) GeneratePassword(ctx context.Context, req *pb.PasswordRequest) (*pb.PasswordResponse, error) {
	length := int(req.Length)
	if length <= 0 {
		length = 16
	}
	if length > 128 {
		length = 128
	}

	count := int(req.Count)
	if count <= 0 {
		count = 1
	}
	if count > 100 {
		count = 100
	}

	// Build character set
	var charset string
	if req.CustomChars != "" {
		charset = req.CustomChars
	} else {
		if req.Lowercase {
			charset += "abcdefghijklmnopqrstuvwxyz"
		}
		if req.Uppercase {
			charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		}
		if req.Numbers {
			charset += "0123456789"
		}
		if req.Symbols {
			charset += "!@#$%^&*()-_=+[]{}|;:,.<>?"
		}
		// Default to all if nothing selected
		if charset == "" {
			charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
		}
	}

	passwords := make([]string, count)
	for i := 0; i < count; i++ {
		password := make([]byte, length)
		for j := 0; j < length; j++ {
			password[j] = charset[gofakeit.Number(0, len(charset)-1)]
		}
		passwords[i] = string(password)
	}

	return &pb.PasswordResponse{Passwords: passwords}, nil
}
