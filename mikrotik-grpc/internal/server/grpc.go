package server

import (
	"log"
	pb "isp-management-system/api/proto"
	"isp-management-system/internal/handler"
	"net"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	handler *handler.MikroTikHandler
}

func New(h *handler.MikroTikHandler) *GRPCServer {
	return &GRPCServer{handler: h}
}

func (s *GRPCServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMikroTikServiceServer(grpcServer, s.handler)

	log.Printf("gRPC server listening at %s", addr)
	return grpcServer.Serve(lis)
}
