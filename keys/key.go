package keys

import (
	"crypto/rsa"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var (
	verifyKey      *rsa.PublicKey
	signKey        *rsa.PrivateKey
	privateKeyPath = "keys/app.rsa"
	publicKeyPath  = "keys/app.rsa.pub"
)

func InitPublicKey() (*rsa.PublicKey, error) {
	if verifyKey == nil {
		verifyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return nil, err
		}
		verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
		if err != nil {
			return nil, err
		}
	}
	return verifyKey, nil
}

func InitPrivateKey() (*rsa.PrivateKey, error) {
	if signKey == nil {
		signBytes, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, err
		}
		signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
		if err != nil {
			return nil, err
		}
	}
	return signKey, nil
}
