package extractor

// Request is a simple type for testing our type extraction
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// Server handles HTTP requests
type Server interface {
	HandleRequest(req Request) Response
	Start() error
	Stop() error
}

// SimpleServer implements the Server interface
type SimpleServer struct {
	Port    int
	Router  map[string]func(Request) Response
	Running bool
}

func (s *SimpleServer) HandleRequest(req Request) Response {
	handler, ok := s.Router[req.Path]
	if !ok {
		return Response{
			StatusCode: 404,
			Body:       []byte("Not Found"),
		}
	}
	return handler(req)
}

func (s *SimpleServer) Start() error {
	s.Running = true
	return nil
}

func (s *SimpleServer) Stop() error {
	s.Running = false
	return nil
}
