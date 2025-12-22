package auth

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	OrgID     string   `json:"org_id"`
	Role      string   `json:"role"`
	Scopes    []string `json:"scopes,omitempty"`
	TokenType string   `json:"typ"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessTokenID    string
	RefreshTokenID   string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

func ParseRSAPrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	priv, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return priv, nil
}

func ParseRSAPublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		pub, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not an RSA public key")
		}
		return pub, nil
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

func GenerateTokenPair(userID, orgID, role string, scopes []string, now time.Time, accessTTL, refreshTTL time.Duration, privateKey *rsa.PrivateKey) (TokenPair, error) {
	accessID := uuid.NewString()
	refreshID := uuid.NewString()
	accessExpires := now.Add(accessTTL)
	refreshExpires := now.Add(refreshTTL)

	accessClaims := Claims{
		OrgID:     orgID,
		Role:      role,
		Scopes:    scopes,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpires),
			ID:        accessID,
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims).SignedString(privateKey)
	if err != nil {
		return TokenPair{}, err
	}

	refreshClaims := Claims{
		OrgID:     orgID,
		Role:      role,
		Scopes:    scopes,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpires),
			ID:        refreshID,
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims).SignedString(privateKey)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessTokenID:    accessID,
		RefreshTokenID:   refreshID,
		AccessExpiresAt:  accessExpires,
		RefreshExpiresAt: refreshExpires,
	}, nil
}

func ParseToken(tokenStr string, publicKey *rsa.PublicKey) (*Claims, error) {
	claims := &Claims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	_, err := parser.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
