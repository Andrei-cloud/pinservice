package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/andrei-cloud/pinservice/pkg/broker"
	"github.com/andrei-cloud/pinservice/pkg/domain"
)

//below block is used for developmentand testing purposes ONLY
const (
	LMK = "DFAEBFECEAFBEFDFAEBFECEAFBEFDFAE" //F4EDC8
	//clear keys
	PVKA = "8D9341F9E728F3DD" //47F249
	PVKB = "9CD1C3BE18DBE869" //49A471
	PVK  = "8D9341F9E728F3DD9CD1C3BE18DBE869"
	TPK  = "FA9F90D49CB27B7D14A3FA9CCCFF6CB7" //804288
	//keys under LMK
	PVK_ENC = "7336D50C47128D710DF450BCB2C6461B"
	TPK_ENC = "C4ED597EE0C9697104ED399BE6F8B872"

	PINBlock = "793AE62DFC8D2426"

	ClearPIN = "1234"
	PVV      = "3843"
	PAN      = "4234070000000102"
)

var (
	ErrInvalidPIN      = errors.New("invalid pin")
	ErrInvalidResponse = errors.New("response is not valid")
	ErrHsmError        = errors.New("hsm error")
)

// PinService describes the service.
type PinService interface {
	Verify(ctx context.Context, pin *domain.PIN) error
	GeneratePVV(ctx context.Context, pin *domain.PIN) (string, error)
}

var _ PinService = &basicPinService{}

type basicPinService struct {
	hsmBroker broker.Broker
}

func (b *basicPinService) Verify(ctx context.Context, pin *domain.PIN) (e0 error) {
	var response []byte

	if b.hsmBroker == nil {
		return fmt.Errorf("hsm broker not initialized")
	}

	command := bytes.Buffer{}
	command.Write([]byte("DCU"))
	command.Write([]byte(TPK_ENC))
	command.Write([]byte(PVK_ENC))
	command.Write([]byte(pin.EncryptedPIN))
	command.Write([]byte("01"))
	command.Write([]byte(pin.PAN)[len(pin.PAN)-13 : len(pin.PAN)-1])
	command.Write([]byte("1"))
	command.Write([]byte(pin.PVV))

	response, e0 = b.hsmBroker.Send(command.Bytes())
	if e0 != nil {
		return e0
	}

	if string(response[:2]) != "DD" {
		e0 = ErrInvalidResponse
	} else {
		code := string(response[2:4])
		if code != "00" {
			e0 = fmt.Errorf(HSMErrors[code], ErrHsmError)
		}
	}

	return e0
}
func (b *basicPinService) GeneratePVV(ctx context.Context, pin *domain.PIN) (s0 string, e1 error) {
	// TODO implement the business logic of GeneratePVV
	return s0, e1
}

// NewBasicPinService returns a naive, stateless implementation of PinService.
func NewBasicPinService(b broker.Broker) PinService {
	return &basicPinService{
		hsmBroker: b,
	}
}

// New returns a PinService with all of the expected middleware wired in.
func New(b broker.Broker, middleware []Middleware) PinService {
	var svc PinService = NewBasicPinService(b)
	for _, m := range middleware {
		svc = m(svc)
	}
	return svc
}
