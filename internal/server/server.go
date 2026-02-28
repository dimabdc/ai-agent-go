package server

// Server represents the HTTP server
type Server struct {
	// dependencies
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start(addr string) error {
	return nil
}
