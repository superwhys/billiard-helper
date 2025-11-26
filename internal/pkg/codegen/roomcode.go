package codegen

import "fmt"

func GenerateRoomCode(id uint) (string, error) {
	code, err := EncodeIDWithMask(uint64(id))
	if err != nil {
		return "", fmt.Errorf("id(%d) generate room code failed: %w", id, err)
	}

	return code, nil
}

func ParseRoomCode(code string) (uint, error) {
	id, err := DecodeIDWithMask(code)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
