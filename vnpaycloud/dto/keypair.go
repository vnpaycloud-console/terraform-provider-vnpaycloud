package dto

// KeyPair matches the iac-proxy-v2 KeyPair proto message.
type KeyPair struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PublicKey   string `json:"publicKey"`
	Fingerprint string `json:"fingerprint"`
	CreatedAt   string `json:"createdAt"`
}

// CreateKeyPairRequest matches the iac-proxy-v2 CreateKeyPairRequest proto message.
type CreateKeyPairRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey,omitempty"`
}

// KeyPairResponse matches the iac-proxy-v2 KeyPairResponse proto message.
type KeyPairResponse struct {
	KeyPair    KeyPair `json:"keyPair"`
	PrivateKey string  `json:"privateKey,omitempty"`
}

// ListKeyPairsResponse matches the iac-proxy-v2 ListKeyPairsResponse proto message.
type ListKeyPairsResponse struct {
	KeyPairs []KeyPair `json:"keyPairs"`
}
