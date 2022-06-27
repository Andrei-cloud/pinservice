// THIS FILE IS AUTO GENERATED BY GK-CLI DO NOT EDIT!!
package http

import (
	endpoint "github.com/andrei-cloud/pinservice/pkg/endpoint"
	http "github.com/go-kit/kit/transport/http"
	http1 "net/http"
)

// NewHTTPHandler returns a handler that makes a set of endpoints available on
// predefined paths.
func NewHTTPHandler(endpoints endpoint.Endpoints, options map[string][]http.ServerOption) http1.Handler {
	m := http1.NewServeMux()
	makeVerifyHandler(m, endpoints, options["Verify"])
	makeGeneratePVVHandler(m, endpoints, options["GeneratePVV"])
	return m
}
