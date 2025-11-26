package codegen

func GenerateRoomCode(id uint) string {
	return EncodeIDWithMask(int64(id))
}

func ParseRoomCode(code string) (uint, error) {
	id, err := DecodeIDWithMask(code)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
