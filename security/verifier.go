package security

import (
	"crypto/ed25519"
	"fmt"
)

type Verifier struct {
	publicKey ed25519.PublicKey
}

func NewVerifier(publicKey ed25519.PublicKey) (*Verifier, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key")
	}

	return &Verifier{
		publicKey: publicKey,
	}, nil
}

func (v *Verifier) Verify(data []byte, signature []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}

	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature")
	}

	if !ed25519.Verify(v.publicKey, data, signature) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}