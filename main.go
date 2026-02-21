package main

import (
	"baize-monitor/internal/server/wire"
	"baize-monitor/pkg/constants"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Initializing server...")
	server, err := wire.InitializeServer(constants.ServerConfigPath)
	if err != nil {
		fmt.Printf("Failed to initialize server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting server...")

	// Start server, exit if startup fails
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All servers started successfully")

	// Set up signal listening for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nReceived shutdown signal, shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Error occurred during server shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server shutdown completed successfully")
}
