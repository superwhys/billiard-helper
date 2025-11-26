package codegen

import (
	"fmt"

	"github.com/miebyte/goutils/logging"
	"github.com/sqids/sqids-go"
)

const (
	alphabet   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	minCodeLen = 8
)

var (
	codeGen *sqids.Sqids
)

func init() {
	var err error
	codeGen, err = sqids.New(sqids.Options{
		MinLength: minCodeLen,
		Alphabet:  alphabet,
	})
	logging.PanicError(err)
}

// 编码
func EncodeIDWithMask(id uint64) (string, error) {
	return codeGen.Encode([]uint64{id})

}

// 解码
func DecodeIDWithMask(code string) (uint64, error) {
	values := codeGen.Decode(code)
	if len(values) == 0 {
		return 0, fmt.Errorf("code(%s) can not decode anything", code)
	}

	return values[0], nil
}
