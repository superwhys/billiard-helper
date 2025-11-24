// File:		middleware.go
// Created by:	Hoven
// Created on:	2025-10-08
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
	"github.com/superwhys/billiard-helper/internal/ports"
)

const (
	TokenContextKey = "token_claims"
)

func TokenVerifyFromSocket(logic ports.TokenAuthLogic) func(r *http.Request) error {
	return func(r *http.Request) error {
		ctx := logging.CloneContext(r.Context())
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			return errcode.ErrCodeNoToken
		}

		_, err := logic.GetUserTokenClaims(ctx, tokenStr)
		if err != nil {
			logging.Errorc(ctx, "get secret from jwt token failed: %v", err)
			if ec, ok := errcode.AsErrcode(err); ok {
				return ec
			}
			return errcode.ErrCodeNoToken
		}

		return nil
	}
}

func TokenVerifyMiddleware(logic ports.TokenAuthLogic) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.GetHeader("Authorization")
		if tokenStr == "" {
			ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrCodeNoToken))
			ctx.Abort()
			return
		}

		claims, err := logic.GetUserTokenClaims(ctx, tokenStr)
		if err != nil {
			logging.Errorc(ctx, "get secret from jwt token failed: %v", err)
			if ec, ok := errcode.AsErrcode(err); ok {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(ec))
			} else {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrCodeNoToken))
			}
			ctx.Abort()
			return
		}

		ctx.Set(string(TokenContextKey), claims)

		logging.Debugc(ctx, "token claims: %s", logging.JsonifyNoIndent(claims))
		// 将 claims 注入 request context，供下游以 context.Value 读取
		reqCtx := context.WithValue(ctx.Request.Context(), TokenContextKey, claims)
		ctx.Request = ctx.Request.WithContext(reqCtx)
	}
}

func TokenClaimsFromContext(ctx context.Context) (*jwt.UserTokenClaims, error) {
	claims, ok := ctx.Value(TokenContextKey).(*jwt.UserTokenClaims)
	if !ok {
		return nil, errcode.ErrCodeNoToken
	}
	return claims, nil
}
