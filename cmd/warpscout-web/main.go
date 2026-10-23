// Command warpscout-web serves the WARPSCOUT scanner through a web UI.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/vernette/warpscout/internal/webserver"
	"github.com/vernette/warpscout/mobileapi"
)

var version = "dev"

type coreRunner struct{}

func (coreRunner) Start(requestJSON string, l webserver.Listener) error {
	return mobileapi.Start(requestJSON, l)
}
func (coreRunner) Cancel() { mobileapi.Cancel() }

func main() {
	listen := flag.String("listen", envOr("WARPSCOUT_LISTEN", "127.0.0.1:8787"), "address to listen on (host:port)")
	dataDir := flag.String("data", envOr("WARPSCOUT_DATA", defaultDataDir()), "directory for the account and scan history")
	token := flag.String("token", os.Getenv("WARPSCOUT_TOKEN"), "access token (default: generated and stored in the data directory)")
	noAuth := flag.Bool("no-auth", false, "disable authentication (only allowed on a loopback address)")
	tlsCert := flag.String("tls-cert", os.Getenv("WARPSCOUT_TLS_CERT"), "TLS certificate file (enables HTTPS together with -tls-key)")
	tlsKey := flag.String("tls-key", os.Getenv("WARPSCOUT_TLS_KEY"), "TLS private key file")
	trustProxy := flag.Bool("trust-proxy", os.Getenv("WARPSCOUT_TRUST_PROXY") == "1", "use X-Forwarded-For for the failed-login limiter (only behind a reverse proxy)")
	demo := flag.Bool("demo", false, "use a fake core, to try the UI without a WARP account or network")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		println(version)
		return
	}
	store, err := webserver.NewStore(*dataDir)
	if err != nil {
		log.Fatalf("data directory: %v", err)
	}

	host, _, err := net.SplitHostPort(*listen)
	if err != nil {
		log.Fatalf("invalid -listen %q: %v", *listen, err)
	}
	loopback := isLoopback(host)

	authToken := *token
	switch {
	case *noAuth:
		if !loopback {
			log.Fatal("-no-auth is only allowed when listening on a loopback address")
		}
		authToken = ""
	case authToken == "":
		authToken, err = loadOrCreateToken(filepath.Join(*dataDir, "token"))
		if err != nil {
			log.Fatalf("token: %v", err)
		}
		log.Printf("access token file: %s", filepath.Join(*dataDir, "token"))
	}

	var runner webserver.Runner = coreRunner{}
	if *demo {
		runner = webserver.NewDemoRunner()
		log.Print("DEMO mode: the core is simulated")
	}
	srv := webserver.New(webserver.Config{Runner: runner, Store: store, Token: authToken, Version: version, TrustProxy: *trustProxy})

	httpServer := &http.Server{
		Addr:              *listen,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	if *tlsCert == "" && !loopback && authToken != "" {
		log.Print("WARNING: serving plain HTTP on a non-loopback address; the token travels unencrypted. Use -tls-cert/-tls-key or a reverse proxy with HTTPS.")
	}

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		mobileapi.Cancel()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	}()

	log.Printf("WARPSCOUT web %s listening on %s", version, *listen)
	if *tlsCert != "" {
		err = httpServer.ListenAndServeTLS(*tlsCert, *tlsKey)
	} else {
		err = httpServer.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func defaultDataDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "warpscout-web")
	}
	return "./warpscout-data"
}

func isLoopback(host string) bool {
	if host == "" {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

func loadOrCreateToken(path string) (string, error) {
	if data, err := os.ReadFile(path); err == nil {
		if t := strings.TrimSpace(string(data)); t != "" {
			return t, nil
		}
	}
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	return token, os.WriteFile(path, []byte(token+"\n"), 0o600)
}
