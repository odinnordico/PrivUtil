package api

// Server holds the implementations of all PrivUtil RPC handlers. The handlers
// are defined as methods across the *_handlers.go files in this package and are
// exposed via the connect adapter in connect_adapter.go.
type Server struct {
	custom *customDict
}

// Option configures a Server.
type Option func(*Server)

// WithCustomDictPath backs the spellchecker's custom dictionary with a file at
// path. An empty path (the default) keeps it in memory only.
func WithCustomDictPath(path string) Option {
	return func(s *Server) { s.custom = newCustomDict(path) }
}

// NewServer builds a Server. By default the custom dictionary is in-memory only;
// pass WithCustomDictPath to persist it.
func NewServer(opts ...Option) *Server {
	s := &Server{custom: newCustomDict("")}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
