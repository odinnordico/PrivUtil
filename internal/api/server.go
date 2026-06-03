package api

// Server holds the implementations of all PrivUtil RPC handlers. The handlers
// are defined as methods across the *_handlers.go files in this package and are
// exposed via the connect adapter in connect_adapter.go.
type Server struct{}

func NewServer() *Server {
	return &Server{}
}
