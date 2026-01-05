package client

import (
	"log"

	"github.com/go-routeros/routeros"
)

type MikroTikClient struct {
	conn *routeros.Client
}

func New(host, user, pass string) (*MikroTikClient, error) {
	log.Printf("Attempting to connect to MikroTik router at %s with user %s", host, user)
	c, err := routeros.Dial(host, user, pass)
	if err != nil {
		log.Printf("Failed to connect to RouterOS: %v", err)
		return nil, err
	}
	log.Println("Successfully connected to MikroTik router")
	return &MikroTikClient{conn: c}, nil
}

func (m *MikroTikClient) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}

func (m *MikroTikClient) Run(args ...string) (*routeros.Reply, error) {
	return m.conn.Run(args...)
}

func (m *MikroTikClient) ActivePPPoE() ([]map[string]string, error) {
	reply, err := m.conn.Run("/ppp/active/print")
	if err != nil {
		return nil, err
	}
	result := make([]map[string]string, len(reply.Re))
	for i, sentence := range reply.Re {
		result[i] = sentence.Map
	}
	return result, nil
}

func (m *MikroTikClient) CreatePPPoEUser(name, password, profile string) error {
	log.Printf("Creating PPPoE user: name=%s, profile=%s", name, profile)
	reply, err := m.conn.Run("/ppp/secret/add",
		"=name="+name, "=password="+password, "=service=pppoe", "=profile="+profile)
	if err != nil {
		log.Printf("Error creating PPPoE user: %v", err)
		return err
	}
	log.Printf("RouterOS reply: %v", reply)
	return nil
}

func (m *MikroTikClient) RemovePPPoEUser(name string) error {
	id, err := m.findPPPoEUserID(name)
	if err != nil {
		return err
	}
	_, err = m.conn.Run("/ppp/secret/remove", "=.id="+id)
	return err
}

func (m *MikroTikClient) findPPPoEUserID(name string) (string, error) {
	reply, err := m.conn.Run("/ppp/secret/print", "?name="+name)
	if err != nil {
		return "", err
	}
	if len(reply.Re) == 0 {
		return "", log.Output(1, "user not found")
	}
	return reply.Re[0].Map[".id"], nil
}

func (m *MikroTikClient) GetPPPoEUser(name string) (map[string]string, error) {
	reply, err := m.conn.Run("/ppp/secret/print", "?name="+name)
	log.Printf("Reply: %v", reply)
	if err != nil {
		return nil, err
	}
	log.Printf("Updating PPPoE user %s with fields: %v", name, reply)

	if len(reply.Re) == 0 {
		return nil, log.Output(1, "user not found")
	}
	return reply.Re[0].Map, nil
}

func (m *MikroTikClient) UpdatePPPoEUser(name string, fields map[string]string) (map[string]string, error) {
	id, err := m.findPPPoEUserID(name)
	if err != nil {
		return nil, err
	}

	args := []string{"/ppp/secret/set", "=.id=" + id}
	for key, value := range fields {
		args = append(args, "="+key+"="+value)
	}

	_, err = m.conn.Run(args...)
	if err != nil {
		return nil, err
	}

	// After updating, fetch the user's new data and return it
	return m.GetPPPoEUser(name)
}
