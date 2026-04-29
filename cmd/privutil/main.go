//go:build manual

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"

	"github.com/odinnordico/privutil/internal/api"
	"github.com/odinnordico/privutil/internal/server"
	pb "github.com/odinnordico/privutil/proto"
)

// Version info (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Define CLI flags
	port := flag.String("port", getEnvOrDefault("PORT", "8090"), "Port to listen on")
	host := flag.String("host", getEnvOrDefault("HOST", ""), "Host to bind to (empty = all interfaces)")
	logLevel := flag.String("log-level", getEnvOrDefault("LOG_LEVEL", "info"), "Log level: debug, info (debug adds file/line to log output)")
	version := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "PrivUtil - Offline-capable developer utility suite\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  PORT       Port to listen on (default: 8090)\n")
		fmt.Fprintf(os.Stderr, "  HOST       Host to bind to (default: all interfaces)\n")
		fmt.Fprintf(os.Stderr, "  LOG_LEVEL  Log level (default: info)\n")
	}

	flag.Parse()

	if *version {
		fmt.Printf("PrivUtil %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	switch *logLevel {
	case "debug":
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case "info":
		// default log flags
	default:
		log.Fatalf("Unsupported log level %q: use debug or info", *logLevel)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterPrivUtilServiceServer(grpcServer, api.NewServer())

	// Create and start HTTP server
	addr := *host + ":" + *port
	srv := server.New(addr, grpcServer)

	log.Printf("Starting PrivUtil on %s...", addr)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
