package endpoint

import (
	"context"

	"github.com/andrei-cloud/pinservice/pkg/domain"
	service "github.com/andrei-cloud/pinservice/pkg/service"
	endpoint "github.com/go-kit/kit/endpoint"
)

// VerifyRequest collects the request parameters for the Verify method.
type VerifyRequest struct {
	*domain.PIN
}

// VerifyResponse collects the response parameters for the Verify method.
type VerifyResponse struct {
	Success bool  `json:"success"`
	E0      error `json:"error"`
}

// MakeVerifyEndpoint returns an endpoint that invokes Verify on the service.
func MakeVerifyEndpoint(s service.PinService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var isSuccess bool
		req := request.(VerifyRequest).PIN
		e0 := s.Verify(ctx, req)
		if e0 == nil {
			isSuccess = true
		}
		return VerifyResponse{Success: isSuccess, E0: e0}, e0
	}
}

// Failed implements Failer.
func (r VerifyResponse) Failed() error {
	return r.E0
}

// GeneratePVVRequest collects the request parameters for the GeneratePVV method.
type GeneratePVVRequest struct {
	*domain.PIN
}

// GeneratePVVResponse collects the response parameters for the GeneratePVV method.
type GeneratePVVResponse struct {
	S0 string `json:"s0"`
	E1 error  `json:"e1"`
}

// MakeGeneratePVVEndpoint returns an endpoint that invokes GeneratePVV on the service.
func MakeGeneratePVVEndpoint(s service.PinService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GeneratePVVRequest).PIN
		s0, e1 := s.GeneratePVV(ctx, req)
		return GeneratePVVResponse{
			E1: e1,
			S0: s0,
		}, nil
	}
}

// Failed implements Failer.
func (r GeneratePVVResponse) Failed() error {
	return r.E1
}

// Failure is an interface that should be implemented by response types.
// Response encoders can check if responses are Failer, and if so they've
// failed, and if so encode them using a separate write path based on the error.
type Failure interface {
	Failed() error
}

// Verify implements Service. Primarily useful in a client.
func (e Endpoints) Verify(ctx context.Context, pin *domain.PIN) (e0 error) {
	response, err := e.VerifyEndpoint(ctx, pin)
	if err != nil {
		return err
	}
	return response.(VerifyResponse).E0
}

// GeneratePVV implements Service. Primarily useful in a client.
func (e Endpoints) GeneratePVV(ctx context.Context, pin *domain.PIN) (s0 string, e1 error) {
	request := GeneratePVVRequest{PIN: pin}
	response, err := e.GeneratePVVEndpoint(ctx, request)
	if err != nil {
		return
	}
	return response.(GeneratePVVResponse).S0, response.(GeneratePVVResponse).E1
}
