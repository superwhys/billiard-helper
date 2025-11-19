package response

import (
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/models/errcode"
)

func ResponseSuccess() *ginutils.Ret[any] {
	return ginutils.SuccessRet[any](nil)
}

func ResponseWithData[T any](data T) *ginutils.Ret[T] {
	return ginutils.SuccessRet(data)
}

func ErrorResponseWithCode(ec errcode.ErrCode) *ginutils.Ret[any] {
	return ginutils.ErrorRet(ec.Code(), ec.String())
}

type PaginatedResponse[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

type TokenResponse struct {
	Token string `json:"token"`
}
