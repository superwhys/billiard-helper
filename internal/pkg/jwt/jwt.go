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
	UserID    uint   `json:"user_id"`
	TokenType string `json:"token_type"`
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

func GenerateTokenPair(signingKey []byte, accessTimeout, refreshTimeout time.Duration, userID uint) (string, string, string, error) {
	if len(signingKey) == 0 {
		return "", "", "", fmt.Errorf("signing key is required")
	}

	sessionID := uuid.NewString()
	now := time.Now()
	accessExpiresAt := now.Add(accessTimeout)
	refreshExpiresAt := now.Add(refreshTimeout)

	accessClaims := &UserTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
		},
		UserID:    userID,
		TokenType: TokenTypeAccess,
	}

	refreshClaims := &UserTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		},
		UserID:    userID,
		TokenType: TokenTypeRefresh,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(signingKey)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(signingKey)
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, sessionID, nil
}

func GenerateAccessToken(signingKey []byte, timeout time.Duration, userID uint, sessionID string) (string, error) {
	if len(signingKey) == 0 {
		return "", fmt.Errorf("signing key is required")
	}
	if sessionID == "" {
		return "", fmt.Errorf("session id is required")
	}

	now := time.Now()
	expiresAt := now.Add(timeout)
	claims := &UserTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID:    userID,
		TokenType: TokenTypeAccess,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return accessToken, nil
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
		return nil, errcode.ErrUnauthorized
	}
	return claims, nil
}

func SetTokenClaimsToContext(ctx context.Context, claims *UserTokenClaims) context.Context {
	return context.WithValue(ctx, TokenContextKey, claims)
}
