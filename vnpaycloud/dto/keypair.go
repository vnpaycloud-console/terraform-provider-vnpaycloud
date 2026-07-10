package dto

// KeyPair matches the backend KeyPair proto message.
type KeyPair struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
	CreatedAt string `json:"createdAt"`
}

// CreateKeyPairRequest matches the backend CreateKeyPairRequest proto message.
type CreateKeyPairRequest struct {
	Name       string `json:"name"`
	PublicKey  string `json:"publicKey,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}

// KeyPairResponse matches the backend KeyPairResponse proto message.
type KeyPairResponse struct {
	KeyPair    KeyPair `json:"keyPair"`
	PrivateKey string  `json:"privateKey,omitempty"`
}

// ListKeyPairsResponse matches the backend ListKeyPairsResponse proto message.
type ListKeyPairsResponse struct {
	KeyPairs []KeyPair `json:"keyPairs"`
}
