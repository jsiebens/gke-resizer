package gkeresizer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
)

// Server is a cleaning server.
type Server struct {
	resizer *Resizer
}

// NewServer creates a new server for handler functions.
func NewServer(resizer *Resizer) (*Server, error) {
	if resizer == nil {
		return nil, fmt.Errorf("missing resizer")
	}

	return &Server{
		resizer: resizer,
	}, nil
}

// HTTPHandler is an http handler that invokes the resizer with the given
// parameters.
func (s *Server) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, status, err := s.clean(r.Body)
		if err != nil {
			s.handleError(w, err, status)
			return
		}

		b, err := json.Marshal(&cleanResp{
			Result: *result,
		})
		if err != nil {
			err = fmt.Errorf("failed to marshal JSON errors: %w", err)
			s.handleError(w, err, 500)
			return
		}

		w.WriteHeader(200)
		w.Header().Set(contentTypeHeader, contentTypeJSON)
		w.Write(b)
	}
}

// clean reads the given body as JSON and starts a resizer instance.
func (s *Server) clean(r io.ReadCloser) (*string, int, error) {
	var p = Payload{
	}
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return nil, 500, fmt.Errorf("failed to decode payload as JSON: %w", err)
	}

	result, err := s.resizer.Resize(p.Project, p.Location, p.Cluster, p.NodePool, p.NodeCount)
	if err != nil {
		return nil, 400, fmt.Errorf("failed to resize: %w", err)
	}

	return result, 200, nil
}

// handleError returns a JSON-formatted error message
func (s *Server) handleError(w http.ResponseWriter, err error, status int) {
	log.Printf("error %d: %s", status, err.Error())

	b, err := json.Marshal(&errorResp{Error: err.Error()})
	if err != nil {
		err = fmt.Errorf("failed to marshal JSON errors: %w", err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(status)
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.Write(b)
}

// Payload is the expected incoming payload format.
type Payload struct {
	// Repo is the name of the repo in the format gcr.io/foo/bar
	Project   string `json:"project"`
	Location  string `json:"location"`
	Cluster   string `json:"cluster"`
	NodePool  string `json:"nodePool"`
	NodeCount int64  `json:"nodeCount"`
}

type cleanResp struct {
	Result string `json:"result"`
}

type errorResp struct {
	Error string `json:"error"`
}
