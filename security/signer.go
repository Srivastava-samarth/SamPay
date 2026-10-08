package security

import (
	"crypto/ed25519"
	"fmt"
)

type Signer struct {
	privateKey ed25519.PrivateKey
}

func NewSigner(privateKey ed25519.PrivateKey) (*Signer, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key")
	}

	return &Signer{
		privateKey: privateKey,
	}, nil
}

func (s *Signer) Sign(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	return ed25519.Sign(s.privateKey, data), nil
}