package domain

type PIN struct {
	RequestId string `json:"-"`

	ClearPIN     int    `json:"clear_pin,omitempty"`
	Length       int    `json:"length,omitempty"`
	PAN          string `json:"pan,omitempty"`
	EncryptedPIN string `json:"encrypted_pin,omitempty"`
	PVV          string `json:"pvv,omitempty"`
}
