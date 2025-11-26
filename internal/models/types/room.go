package types

import (
	"fmt"
	"strconv"
	"strings"
)

type RoomStatus uint

const (
	RoomStatusPending    RoomStatus = iota + 1 // 未开始
	RoomStatusInProgress                       // 进行中
	RoomStatusFinished                         // 已完成
)

type Room struct {
	ID     uint       `json:"id"`
	UserID uint       `json:"user_id"`
	Status RoomStatus `json:"status"`

	Players []*Player `json:"players,omitempty"`
	Scores  []*Scores `json:"scores,omitempty"`
}

func IsSocketRoomID(roomID string) bool {
	return strings.HasPrefix(roomID, "room_")
}

func SocketRoomID(roomID uint) string {
	return fmt.Sprintf("room_%d", roomID)
}

func ParseRoomID(roomID string) uint {
	if !IsSocketRoomID(roomID) {
		return 0
	}
	id, err := strconv.ParseUint(strings.TrimPrefix(roomID, "room_"), 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}
