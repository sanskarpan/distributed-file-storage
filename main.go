package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anthdm/foreverstore/config"
	"github.com/anthdm/foreverstore/logger"
	"github.com/anthdm/foreverstore/p2p"
)

func makeServer(cfg *config.Config) *FileServer {
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    cfg.ListenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	var encKey []byte
	if cfg.EncryptionEnabled {
		if cfg.EncryptionKey != "" {
			// TODO: Parse hex key from config
			encKey = newEncryptionKey()
		} else {
			encKey = newEncryptionKey()
		}
	}

	fileServerOpts := FileServerOpts{
		EncKey:            encKey,
		StorageRoot:       cfg.StorageRoot,
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    cfg.BootstrapNodes,
	}

	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logging
	logger.SetGlobalLevel(cfg.GetLogLevel())
	if cfg.LogFile != "" {
		logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		defer logFile.Close()
		logger.SetGlobalOutput(logFile)
	}

	logger.Info("Starting distributed file storage system")
	logger.Info("Configuration: Listen=%s, Storage=%s, Encryption=%v", 
		cfg.ListenAddr, cfg.StorageRoot, cfg.EncryptionEnabled)

	// Create and start the file server
	server := makeServer(cfg)
	
	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("Server failed to start: %v", err)
		}
	}()

	// Run demo if this is a test setup
	if cfg.ListenAddr == ":3000" {
		runDemo()
	}

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Received shutdown signal, stopping server...")
	server.Stop()
	logger.Info("Server stopped gracefully")
}

func runDemo() {
	// Wait a bit for server to start
	time.Sleep(2 * time.Second)
	
	logger.Info("Running demo...")
	
	// This is a simple demo - in a real application, you'd use the CLI or API
	// For now, we'll just log that the demo would run
	logger.Info("Demo completed - in a real setup, use the CLI or API to interact with the system")
}
