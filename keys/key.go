package keys

import (
	"context"
	"crypto/rsa"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/roysitumorang/laukpauk/helper"
	"go.uber.org/zap"
)

var (
	verifyKey      *rsa.PublicKey
	signKey        *rsa.PrivateKey
	privateKeyPath = "keys/app.rsa"
	publicKeyPath  = "keys/app.rsa.pub"
)

func InitPublicKey() (*rsa.PublicKey, error) {
	ctx := context.Background()
	ctxt := "Keys-InitPublicKey"
	if verifyKey == nil {
		verifyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrReadFile")
			return nil, err
		}
		verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrParseRSAPublicKeyFromPEM")
			return nil, err
		}
	}
	return verifyKey, nil
}

func InitPrivateKey() (*rsa.PrivateKey, error) {
	ctx := context.Background()
	ctxt := "Keys-InitPrivateKey"
	if signKey == nil {
		signBytes, err := os.ReadFile(privateKeyPath)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrReadFile")
			return nil, err
		}
		signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrParseRSAPublicKeyFromPEM")
			return nil, err
		}
	}
	return signKey, nil
}
