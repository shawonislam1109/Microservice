package service

import (
	"github.com/yourname/mikrotik-grpc/internal/client"
)

type MikrotikService struct {
	rosClient *client.MikroTikClient
}

func NewMikrotikService(rosClient *client.MikroTikClient) *MikrotikService {
	return &MikrotikService{
		rosClient: rosClient,
	}
}

func (s *MikrotikService) GetActivePPPoE() ([]map[string]string, error) {
	return s.rosClient.ActivePPPoE()
}

func (s *MikrotikService) CreatePPPoEUser(name, password, profile string) error {
	return s.rosClient.CreatePPPoEUser(name, password, profile)
}

func (s *MikrotikService) RemovePPPoEUser(name string) error {
	return s.rosClient.RemovePPPoEUser(name)
}

func (s *MikrotikService) GetPPPoEUser(name string) (map[string]string, error) {
	return s.rosClient.GetPPPoEUser(name)
}

func (s *MikrotikService) UpdatePPPoEUser(name string, fields map[string]string) (map[string]string, error) {
	return s.rosClient.UpdatePPPoEUser(name, fields)
}