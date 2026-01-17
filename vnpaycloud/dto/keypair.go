package dto

type KeyPair struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key"`
	UserID      string `json:"user_id"`
	Type        string `json:"type"`
}

type CreateKeyPairOpts struct {
	Name       string            `json:"name" required:"true"`
	UserID     string            `json:"user_id,omitempty"`
	Type       string            `json:"type,omitempty"`
	PublicKey  string            `json:"public_key,omitempty"`
	ValueSpecs map[string]string `json:"value_specs,omitempty"`
}

type CreateKeyPairRequest struct {
	KeyPair CreateKeyPairOpts `json:"keypair"`
}

type CreateKeyPairResponse struct {
	KeyPair KeyPair `json:"keypair"`
}

type DeleteKeyPairResponse struct {
}

type GetKeyPairResponse struct {
	KeyPair KeyPair `json:"keypair"`
}

type GetKeyPairOpts struct {
	UserID string `q:"user_id"`
}

type DeleteKeyPairOpts struct {
	UserID string `q:"user_id"`
}
