package codegen

import (
	"errors"
	"strings"
)

const (
	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base     = 36
	codeLen  = 8        // 固定长度8位
	mask     = 87231231 // 任意正整数（建议换成你自己的！）
)

// 编码
func EncodeIDWithMask(id int64) string {
	codeNum := id ^ mask
	var sb strings.Builder
	if codeNum == 0 {
		return strings.Repeat("A", codeLen)
	}
	for codeNum > 0 {
		rem := codeNum % base
		sb.WriteByte(alphabet[rem])
		codeNum /= base
	}
	code := sb.String()
	// 翻转
	buf := []rune(code)
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	realCode := string(buf)
	// 左侧补A
	padLen := codeLen - len(realCode)
	if padLen > 0 {
		realCode = strings.Repeat("A", padLen) + realCode
	}
	return realCode
}

// 解码
func DecodeIDWithMask(code string) (int64, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != codeLen {
		return 0, errors.New("invalid code length")
	}
	code = strings.TrimLeft(code, "A")
	if code == "" {
		return 0, nil
	}
	var codeNum int64
	for _, c := range code {
		idx := strings.IndexRune(alphabet, c)
		if idx < 0 {
			return 0, errors.New("invalid code value")
		}
		codeNum = codeNum*base + int64(idx)
	}
	id := codeNum ^ mask
	return id, nil
}
