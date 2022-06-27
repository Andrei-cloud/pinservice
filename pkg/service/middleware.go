package service

import (
	"context"

	"github.com/andrei-cloud/pinservice/pkg/domain"
	log "github.com/go-kit/log"
)

// Middleware describes a service middleware.
type Middleware func(PinService) PinService

type loggingMiddleware struct {
	logger log.Logger
	next   PinService
}

// LoggingMiddleware takes a logger as a dependency
// and returns a PinService Middleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next PinService) PinService {
		return &loggingMiddleware{logger, next}
	}

}

func (l loggingMiddleware) Verify(ctx context.Context, pin *domain.PIN) (e0 error) {
	defer func() {
		l.logger.Log("method", "Verify", "request", pin.RequestId, "err", e0)
	}()
	return l.next.Verify(ctx, pin)
}

func (l loggingMiddleware) GeneratePVV(ctx context.Context, pin *domain.PIN) (s0 string, e1 error) {
	defer func() {
		l.logger.Log("method", "GeneratePVV", "request", pin, "s0", s0, "e1", e1)
	}()
	return l.next.GeneratePVV(ctx, pin)
}
