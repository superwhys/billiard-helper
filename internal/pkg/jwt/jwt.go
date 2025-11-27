package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/superwhys/billiard-helper/internal/errcode"
)

var ErrTokenExpired = jwt.ErrTokenExpired

type UserTokenClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"user_id"`
}

func GenerateToken(signingKey []byte, timeout time.Duration, userID uint) (string, error) {
	if len(signingKey) == 0 {
		return "", fmt.Errorf("signing key is required")
	}

	expiresAt := time.Now().Add(timeout)
	claims := &UserTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.NewString(), // JTI
		},
		UserID: userID,
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return token, nil
}

func ParseToken(tokenStr string, signingKey []byte) (*UserTokenClaims, error) {
	tc := &UserTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, tc, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return tc, nil
}

type ContextKey string

const TokenContextKey ContextKey = "user_token_claims"

func TokenClaimsFromContext(ctx context.Context) (*UserTokenClaims, error) {
	claims, ok := ctx.Value(TokenContextKey).(*UserTokenClaims)
	if !ok {
		return nil, errcode.ErrCodeNoToken
	}
	return claims, nil
}
