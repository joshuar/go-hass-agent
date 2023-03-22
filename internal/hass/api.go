package hass

//go:generate go-enum --marshal

// ENUM(encrypted,get_config,update_location)
type RequestType string

type GenericRequest struct {
	Type          RequestType `json:"type"`
	Data          interface{} `json:"data,omitempty"`
	Encrypted     bool        `json:"encrypted,omitempty"`
	EncryptedData interface{} `json:"encrypted_data,omitempty"`
}
