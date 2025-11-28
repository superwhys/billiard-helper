package response

import (
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/internal/errcode"
)

func ResponseSuccess() *ginutils.Ret[any] {
	return ginutils.SuccessRet[any](nil)
}

func ResponseWithData[T any](data T) *ginutils.Ret[T] {
	return ginutils.SuccessRet(data)
}

func ErrorResponseWithCode(ec errcode.ErrCode) *ginutils.Ret[any] {
	return ginutils.ErrorRet(ec)
}
