package handler

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/yourname/mikrotik-grpc/api/proto"
	"github.com/yourname/mikrotik-grpc/internal/client"
	"github.com/yourname/mikrotik-grpc/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	idleTimeout = 30 * time.Second
	cleanupInterval = 10 * time.Second
)

type cachedClient struct {
	client   *client.MikroTikClient
	lastUsed time.Time
}

type MikroTikHandler struct {
	pb.UnimplementedMikroTikServiceServer
	clientCache map[string]*cachedClient
	mu          sync.Mutex
}

func New() *MikroTikHandler {
	h := &MikroTikHandler{
		clientCache: make(map[string]*cachedClient),
	}
	go h.startCleanupRoutine()
	return h
}

func (h *MikroTikHandler) getClient(conn *pb.Connection) (*client.MikroTikClient, error) {
	if conn == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Connection details are required")
	}

	key := fmt.Sprintf("%s-%s", conn.Host, conn.User)

	h.mu.Lock()
	defer h.mu.Unlock()

	if cached, ok := h.clientCache[key]; ok {
		// Simple check if client is alive, might need a proper health check
		// For now, we assume if it's in the cache, it's good.
		cached.lastUsed = time.Now()
		return cached.client, nil
	}

	mtClient, err := client.New(conn.Host, conn.User, conn.Pass)
	if err != nil {
		return nil, err
	}

	h.clientCache[key] = &cachedClient{
		client:   mtClient,
		lastUsed: time.Now(),
	}

	return mtClient, nil
}

func (h *MikroTikHandler) startCleanupRoutine() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		for key, cached := range h.clientCache {
			if time.Since(cached.lastUsed) > idleTimeout {
				cached.client.Close()
				delete(h.clientCache, key)
			}
		}
		h.mu.Unlock()
	}
}

func (h *MikroTikHandler) GetPPPoEActive(ctx context.Context, req *pb.GetPPPoEActiveRequest) (*pb.PPPoEUsers, error) {
	mtClient, err := h.getClient(req.Conn)
	if err != nil {
		return nil, err
	}

	svc := service.NewMikrotikService(mtClient)
	users, err := svc.GetActivePPPoE()
	if err != nil {
		return nil, err
	}

	resp := &pb.PPPoEUsers{}
	for _, u := range users {
		resp.Users = append(resp.Users, &pb.PPPoEUser{
			Attributes: u,
		})
	}
	return resp, nil
}

func (h *MikroTikHandler) CreatePPPoEUser(ctx context.Context, req *pb.CreatePPPoEUserRequest) (*pb.Response, error) {
	if req.Name == "" {
		return &pb.Response{Success: false, Message: "User 'name' is required"}, nil
	}

	mtClient, err := h.getClient(req.Conn)
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}

	svc := service.NewMikrotikService(mtClient)
	err = svc.CreatePPPoEUser(req.Name, req.Password, req.Profile)
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	return &pb.Response{Success: true, Message: "User created"}, nil
}

func (h *MikroTikHandler) RemovePPPoEUser(ctx context.Context, req *pb.RemovePPPoEUserRequest) (*pb.Response, error) {
	mtClient, err := h.getClient(req.Conn)
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}

	svc := service.NewMikrotikService(mtClient)
	err = svc.RemovePPPoEUser(req.Name)
	if err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	return &pb.Response{Success: true, Message: "User removed"}, nil
}

func (h *MikroTikHandler) GetPPPoEUser(ctx context.Context, req *pb.GetPPPoEUserRequest) (*pb.PPPoEUser, error) {
	mtClient, err := h.getClient(req.Conn)
	if err != nil {
		return nil, err
	}

	svc := service.NewMikrotikService(mtClient)
	user, err := svc.GetPPPoEUser(req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.PPPoEUser{
		Attributes: user,
	}, nil
}

func (h *MikroTikHandler) UpdatePPPoEUser(ctx context.Context, req *pb.UpdatePPPoEUserRequest) (*pb.PPPoEUser, error) {
	if len(req.FieldsToUpdate) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "At least one field to update is required")
	}

	mtClient, err := h.getClient(req.Conn)
	if err != nil {
		return nil, err
	}

	svc := service.NewMikrotikService(mtClient)
	updatedUser, err := svc.UpdatePPPoEUser(req.Name, req.FieldsToUpdate)
	if err != nil {
		return nil, err
	}
	return &pb.PPPoEUser{
		Attributes: updatedUser,
	}, nil
}

func (h *MikroTikHandler) CheckConnection(ctx context.Context, req *pb.CheckConnectionRequest) (*pb.Response, error) {
	mtClient, err := h.getClient(req.Conn)
	if err != nil {
		return &pb.Response{Success: false, Message: "Connection failed: " + err.Error()}, nil
	}

	// The connection is good, but we don't want to close it here anymore.
	// The cleanup routine will handle it.
	// We can perhaps run a simple command to be 100% sure.
	svc := service.NewMikrotikService(mtClient)
	_, err = svc.GetActivePPPoE() // A simple read command to check liveness
	if err != nil {
		// If this fails, we should probably remove the client from the cache
		key := fmt.Sprintf("%s-%s", req.Conn.Host, req.Conn.User)
		h.mu.Lock()
		if cached, ok := h.clientCache[key]; ok {
			cached.client.Close()
			delete(h.clientCache, key)
		}
		h.mu.Unlock()
		return &pb.Response{Success: false, Message: "Connection check failed: " + err.Error()}, nil
	}

	return &pb.Response{Success: true, Message: "Connection successful"}, nil
}