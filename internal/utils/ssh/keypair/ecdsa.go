package keypair

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

type ECDSAKeyPair struct {
	PrivateKey string
	PublicKey  string
}

// generatePrivateKey creates a RSA Private Key of specified byte size.
func NewECDSA() (*ECDSAKeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	encoded, err := encodePrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, err
	}

	publicKey, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return &ECDSAKeyPair{
		PrivateKey: string(encoded),
		PublicKey:  string(publicKey),
	}, nil
}

func generatePublicKey(privateKey *ecdsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privateKey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	return pubKeyBytes, nil
}

func encodePrivateKeyToPEM(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	privDER, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return nil, err
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(privDER)

	return privatePEM, nil
}
