// File:		middleware.go
// Created by:	Hoven
// Created on:	2025-10-08
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/models/constant"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/response"
	"github.com/superwhys/billiard-helper/pkg/jwt"
	"github.com/superwhys/billiard-helper/ports"
)

const (
	TokenContextKey = "token_claims"
)

func TokenVerifyFromSocket(logic ports.TokenAuthLogic) func(s websocketutils.Socket) error {
	return func(s websocketutils.Socket) error {
		ctx := s.Context()
		tokenStr := s.Request().Header.Get("Authorization")
		if tokenStr == "" {
			s.Emit(constant.EventClientAuthFailed, response.ErrorResponseWithCode(errcode.ErrCodeNoToken))
			return errcode.ErrCodeNoToken
		}

		claims, err := logic.GetUserTokenClaims(ctx, tokenStr)
		if err != nil {
			logging.Errorc(ctx, "get secret from jwt token failed: %v", err)
			if ec, ok := errcode.AsErrcode(err); ok {
				s.Emit(constant.EventClientAuthFailed, response.ErrorResponseWithCode(ec))
			} else {
				s.Emit(constant.EventClientAuthFailed, response.ErrorResponseWithCode(errcode.ErrCodeNoToken))
			}
			return err
		}

		ctx = context.WithValue(ctx, TokenContextKey, claims)
		s.SetContext(ctx)

		return s.Emit(constant.EventClientAuthSuccess, response.ResponseWithData(claims.User))
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
