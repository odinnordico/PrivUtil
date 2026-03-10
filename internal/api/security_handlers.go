package api

import (
	"bytes"
	"context"
	"crypto/md5"  // #nosec G501
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" // #nosec G505
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	pb "github.com/odinnordico/privutil/proto"
)

func (s *Server) CalculateHash(ctx context.Context, req *pb.HashRequest) (*pb.HashResponse, error) {
	var hash string
	data := []byte(req.Text)

	switch req.Algo {
	case "md5":
		// #nosec G401 G501 - MD5 is intentionally provided as a utility feature
		sum := md5.Sum(data)
		hash = hex.EncodeToString(sum[:])
	case "sha1":
		// #nosec G401 G505 - SHA1 is intentionally provided as a utility feature
		sum := sha1.Sum(data)
		hash = hex.EncodeToString(sum[:])
	case "sha512":
		sum := sha512.Sum512(data)
		hash = hex.EncodeToString(sum[:])
	case "bcrypt":
		cost := bcrypt.DefaultCost
		if reqCost := req.GetCost(); int(reqCost) >= bcrypt.MinCost && int(reqCost) <= bcrypt.MaxCost {
			cost = int(reqCost)
		}
		hashBytes, err := bcrypt.GenerateFromPassword(data, cost)
		if err != nil {
			return nil, fmt.Errorf("failed to generate bcrypt hash: %w", err)
		}
		hash = string(hashBytes)
	default:
		sum := sha256.Sum256(data)
		hash = hex.EncodeToString(sum[:])
	}

	return &pb.HashResponse{Hash: hash}, nil
}

func (s *Server) JwtDecode(ctx context.Context, req *pb.JwtRequest) (*pb.JwtResponse, error) {
	parts := strings.Split(req.Token, ".")
	if len(parts) < 2 {
		return &pb.JwtResponse{Error: "Invalid JWT format"}, nil
	}

	decodeSegment := func(seg string) string {
		if l := len(seg) % 4; l > 0 {
			seg += strings.Repeat("=", 4-l)
		}
		b, err := base64.URLEncoding.DecodeString(seg)
		if err != nil {
			return fmt.Sprintf("Error decoding: %v", err)
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, b, "", "  "); err == nil {
			return pretty.String()
		}
		return string(b)
	}

	return &pb.JwtResponse{
		Header:  decodeSegment(parts[0]),
		Payload: decodeSegment(parts[1]),
	}, nil
}

func (s *Server) CertParse(ctx context.Context, req *pb.CertRequest) (*pb.CertResponse, error) {
	block, _ := pem.Decode([]byte(req.Data))
	if block == nil {
		return &pb.CertResponse{Error: "Failed to decode PEM block"}, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return &pb.CertResponse{Error: fmt.Sprintf("Failed to parse certificate: %v", err)}, nil
	}

	return &pb.CertResponse{
		Subject:   cert.Subject.String(),
		Issuer:    cert.Issuer.String(),
		NotBefore: cert.NotBefore.Format("2006-01-02T15:04:05Z"), // Simplified format or use time.RFC3339 if imported
		NotAfter:  cert.NotAfter.Format("2006-01-02T15:04:05Z"),
		Sans:      cert.DNSNames,
	}, nil
}

func (s *Server) GenerateRsaKeyPair(ctx context.Context, req *pb.RsaKeyRequest) (*pb.RsaKeyResponse, error) {
	bits := int(req.GetBits())
	if bits == 0 {
		bits = 2048
	}
	if bits < 1024 || bits > 8192 {
		return &pb.RsaKeyResponse{Error: "bits must be between 1024 and 8192"}, nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return &pb.RsaKeyResponse{Error: fmt.Sprintf("failed to generate RSA key: %v", err)}, nil
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return &pb.RsaKeyResponse{Error: fmt.Sprintf("failed to marshal RSA public key: %v", err)}, nil
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return &pb.RsaKeyResponse{
		PrivateKey: string(privPEM),
		PublicKey:  string(pubPEM),
	}, nil
}
