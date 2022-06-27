package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	endpoint "github.com/andrei-cloud/pinservice/pkg/endpoint"
	"github.com/andrei-cloud/pinservice/pkg/service"
	http1 "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
)

// makeVerifyHandler creates the handler logic
func makeVerifyHandler(m *http.ServeMux, endpoints endpoint.Endpoints, options []http1.ServerOption) {
	m.Handle("/verify", http1.NewServer(endpoints.VerifyEndpoint, decodeVerifyRequest, encodeVerifyResponse, options...))
}

// decodeVerifyRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeVerifyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := endpoint.VerifyRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if req.RequestId = r.Header.Get("Request-ID"); req.RequestId == "" {
		req.RequestId = uuid.New().String()
	}
	return req, err
}

// encodeVerifyResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer
func encodeVerifyResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(endpoint.Failure); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	return
}

// makeGeneratePVVHandler creates the handler logic
func makeGeneratePVVHandler(m *http.ServeMux, endpoints endpoint.Endpoints, options []http1.ServerOption) {
	m.Handle("/generate-pvv", http1.NewServer(endpoints.GeneratePVVEndpoint, decodeGeneratePVVRequest, encodeGeneratePVVResponse, options...))
}

// decodeGeneratePVVRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeGeneratePVVRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := endpoint.GeneratePVVRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// encodeGeneratePVVResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer
func encodeGeneratePVVResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(endpoint.Failure); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	return
}
func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Success: false, Error: err.Error()})
}
func ErrorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

// This is used to set the http status, see an example here :
// https://github.com/go-kit/kit/blob/master/examples/addsvc/pkg/addtransport/http.go#L133
func err2code(err error) int {
	if errors.Is(err, service.ErrHsmError) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

type errorWrapper struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
