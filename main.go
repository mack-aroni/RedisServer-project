package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
)

// logMessage logs messages with the appropriate level, message, and key-value pairs
func logMessage(level slog.Level, message string, args ...interface{}) {
	logger := slog.Default() // Use the default logger.
	logger.Log(context.Background(), level, message, args...)
}

// Main Server Exec Code
func main() {
	listenAddr := flag.String("listenAddr", defaultListenAddr, "listen address of goredis server")
	flag.Parse()

	server := NewServer(ServerConfig{
		ListenAddr: *listenAddr,
	})

	defer server.Shutdown()

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
