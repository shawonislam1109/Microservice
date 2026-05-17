package main

import (
	"log"
	"isp-management-system/internal/handler"
	"isp-management-system/internal/server"
)

func main() {
	h := handler.New()
	grpcServer := server.New(h)

	log.Println("Starting gRPC server on :50051")
	if err := grpcServer.Start(":50051"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
