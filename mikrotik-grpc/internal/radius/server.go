package radius

import (
	"fmt"
	"log"
	"net"

	"layeh.com/radius"
)

// Server represents the RADIUS server.
type Server struct {
	authHandler radius.Handler
	acctHandler radius.Handler
	secret      string
	authPort    int
	acctPort    int
}

// NewServer creates a new RADIUS server with the given handlers and configuration.
func NewServer(authHandler, acctHandler radius.Handler, secret string, authPort, acctPort int) *Server {
	return &Server{
		authHandler: authHandler,
		acctHandler: acctHandler,
		secret:      secret,
		authPort:    authPort,
		acctPort:    acctPort,
	}
}

// ListenAndServe starts the RADIUS server and blocks until an error occurs.
func (s *Server) ListenAndServe() error {
	errChan := make(chan error, 2)

	// Start authentication server
	go func() {
		authAddr := fmt.Sprintf(":%d", s.authPort)
		log.Printf("RADIUS Server: Starting authentication listener on %s", authAddr)
		server := radius.NewServer(s.authHandler, radius.Secret(s.secret))
		if err := server.ListenAndServe(authAddr); err != nil {
			errChan <- fmt.Errorf("auth server failed: %w", err)
		}
	}()

	// Start accounting server
	go func() {
		acctAddr := fmt.Sprintf(":%d", s.acctPort)
		log.Printf("RADIUS Server: Starting accounting listener on %s", acctAddr)
		server := radius.NewServer(s.acctHandler, radius.Secret(s.secret))
		if err := server.ListenAndServe(acctAddr); err != nil {
			errChan <- fmt.Errorf("acct server failed: %w", err)
		}
	}()

	// Block until one of the servers fails
	return <-errChan
}

// GetLocalIP is a helper function to find a non-loopback local IP address.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}