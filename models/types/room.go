package types

import "fmt"

type RoomStatus uint

const (
	RoomStatusPending    RoomStatus = iota + 1 // 未开始
	RoomStatusInProgress                       // 进行中
	RoomStatusFinished                         // 已完成
)

type Room struct {
	ID       uint       `json:"id"`
	RoomCode string     `json:"room_code"`
	UserID   uint       `json:"user_id"`
	Status   RoomStatus `json:"status"`

	Players []*Player `json:"players,omitempty"`
	Scores  []*Scores `json:"scores,omitempty"`
}

func (r *Room) SocketRoomID() string {
	return fmt.Sprintf("room_%s", r.RoomCode)
}
