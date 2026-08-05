//go:build manual

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	connect "connectrpc.com/connect"

	"github.com/odinnordico/privutil/internal/api"
	"github.com/odinnordico/privutil/internal/server"
	protoconnect "github.com/odinnordico/privutil/proto/protoconnect"
)

// Version info (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// maxRPCRequestBytes caps the size of any single RPC request body. 32 MiB is
// generously above the media handlers' own 10 MB limits while still preventing
// unbounded allocations from a hostile client.
const maxRPCRequestBytes = 32 << 20

func main() {
	// Define CLI flags
	port := flag.String("port", getEnvOrDefault("PORT", "8090"), "Port to listen on")
	host := flag.String("host", getEnvOrDefault("HOST", "127.0.0.1"), "Host to bind to (default: loopback; use 0.0.0.0 for all interfaces)")
	allowedHosts := flag.String("allowed-hosts", getEnvOrDefault("ALLOWED_HOSTS", ""), "Comma-separated extra Host header values to accept (for LAN/custom-domain access)")
	customDict := flag.String("custom-dict", getEnvOrDefault("CUSTOM_DICT", defaultCustomDictPath()), "Path to the spellchecker custom-dictionary file (empty disables persistence)")
	logLevel := flag.String("log-level", getEnvOrDefault("LOG_LEVEL", "info"), "Log level: debug, info (debug adds file/line to log output)")
	version := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "PrivUtil - Offline-capable developer utility suite\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  PORT           Port to listen on (default: 8090)\n")
		fmt.Fprintf(os.Stderr, "  HOST           Host to bind to (default: 127.0.0.1; use 0.0.0.0 for all interfaces)\n")
		fmt.Fprintf(os.Stderr, "  ALLOWED_HOSTS  Extra Host header values to accept, comma-separated\n")
		fmt.Fprintf(os.Stderr, "  CUSTOM_DICT    Path to the spellchecker custom-dictionary file\n")
		fmt.Fprintf(os.Stderr, "  LOG_LEVEL      Log level (default: info)\n")
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

	// Build the connect handler over the existing handlers, with panic recovery.
	// WithReadMaxBytes bounds every RPC body (default is unlimited), capping the
	// memory/CPU amplification available to abusive requests. The media handlers
	// enforce their own stricter 10 MB limits on top of this.
	connectSrv := api.NewConnectServer(api.NewServer(api.WithCustomDictPath(*customDict)))
	rpcPath, rpcHandler := protoconnect.NewPrivUtilServiceHandler(
		connectSrv,
		connect.WithInterceptors(api.RecoveryInterceptor(*logLevel == "debug")),
		connect.WithReadMaxBytes(maxRPCRequestBytes),
	)

	// Create and start HTTP server
	addr := *host + ":" + *port
	srv := server.New(addr, rpcPath, rpcHandler, buildAllowedHosts(*host, *allowedHosts))

	log.Printf("Starting PrivUtil on %s...", addr)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// buildAllowedHosts returns the Host-header allowlist: the loopback names
// always, the concrete bind host when it isn't a wildcard, plus any extra hosts
// (comma-separated). This is what lets the container (HOST=0.0.0.0) still be
// reached at localhost while blocking DNS-rebinding from other hostnames.
func buildAllowedHosts(bindHost, extra string) []string {
	hosts := []string{"localhost", "127.0.0.1", "::1"}
	if bindHost != "" && bindHost != "0.0.0.0" && bindHost != "::" {
		hosts = append(hosts, bindHost)
	}
	for _, h := range strings.Split(extra, ",") {
		if h = strings.TrimSpace(h); h != "" {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

// defaultCustomDictPath returns the per-user location for the spellchecker's
// custom dictionary (e.g. ~/.config/privutil/custom-dictionary.txt), or "" when
// the OS config dir can't be determined — in which case the dictionary is
// in-memory only until a path is set via -custom-dict / CUSTOM_DICT.
func defaultCustomDictPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "privutil", "custom-dictionary.txt")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
