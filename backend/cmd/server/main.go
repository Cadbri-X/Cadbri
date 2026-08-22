package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cadbri/internal/api"
	"cadbri/internal/config"
	"cadbri/internal/engine"
	"cadbri/internal/engine/engines"
	"cadbri/internal/network"
)

func main() {
	configPath := flag.String("config", "", "Path to settings.yml configuration file")
	portFlag := flag.Int("port", 0, "HTTP port to bind (default from settings.yml or 2222)")
	flag.Parse()

	log.Println("==========================================================")
	log.Println("        Cadbri - Ultra-Fast Search Aggregator             ")
	log.Println("==========================================================")

	// 1. Load Settings
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Printf("[Warning] Failed loading settings.yml: %v. Using defaults.", err)
		cfg = config.DefaultConfig()
	}

	port := cfg.Server.Port
	if *portFlag > 0 {
		port = *portFlag
	}
	if port == 0 {
		port = 2222
	}

	// 2. Register Search Engines
	log.Println("[Engines] Initializing engine registry...")
	engines.RegisterAll()
	allEngines := engine.GlobalRegistry.GetAll()
	log.Printf("[Engines] Successfully loaded and registered %d search engines across 7 categories.\n", len(allEngines))

	// 3. Initialize High-Performance Network Client Pool
	clientOpts := network.ClientOptions{
		Timeout:         time.Duration(cfg.Outgoing.RequestTimeout * float64(time.Second)),
		MaxIdleConns:    cfg.Outgoing.PoolConnections,
		MaxConnsPerHost: cfg.Outgoing.PoolMaxSize,
		InsecureSkipTLS: !cfg.Outgoing.Verify,
	}
	if proxy, ok := cfg.Outgoing.Proxies["http"]; ok && proxy != "" {
		clientOpts.ProxyURL = proxy
	} else if proxy, ok := cfg.Outgoing.Proxies["https"]; ok && proxy != "" {
		clientOpts.ProxyURL = proxy
	}

	netClient := network.NewClient(clientOpts)

	// 4. Initialize HTTP Server
	server := api.NewServer(cfg, netClient, engine.GlobalRegistry)
	bindAddr := cfg.Server.BindAddress
	if envBind := os.Getenv("CADBRI_BIND_ADDRESS"); envBind != "" {
		bindAddr = envBind
	}
	if envPort := os.Getenv("CADBRI_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			port = p
		}
	}
	if bindAddr == "" || bindAddr == "127.0.0.1" {
		bindAddr = "0.0.0.0"
	}

	addr := fmt.Sprintf("%s:%d", bindAddr, port)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start Server in Background
	go func() {
		log.Printf("[Server] Cadbri Go backend listening on http://%s\n", addr)
		log.Printf("[Server] JSON Search API: http://localhost:%d/search?q=test&format=json\n", port)
		log.Printf("[Server] Autocompleter API: http://localhost:%d/autocompleter?q=test\n", port)
		log.Printf("[Server] Health check: http://localhost:%d/healthz\n", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Server] Fatal error: %v\n", err)
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Server] Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("[Server] Forced shutdown: %v\n", err)
	}

	log.Println("[Server] Cadbri exited cleanly.")
}
