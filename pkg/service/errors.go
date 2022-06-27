package service

var HSMErrors = map[string]string{
	"01": "pin verification failure: %w",
	"10": "tpk parity error: %w",
	"11": "pvk parity error: %w",
	"27": "pvk not double length: %w",
	"68": "command disabled: %w",
	"69": "pin block format has been disabled: %w",
}
