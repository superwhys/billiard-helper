package match

import "time"

// Room 聚合根
type Room struct {
	ID        uint        `json:"id"`
	RoomCode  string      `json:"room_code"`
	OwnerID   uint        `json:"owner_id"`
	Status    RoomStatus  `json:"status"`
	Config    *GameConfig `json:"config"`
	Players   []*Player   `json:"players"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type Player struct {
	ID       uint       `json:"id"`
	RoomID   uint       `json:"room_id"`
	UserID   *uint      `json:"user_id"`
	NickName string     `json:"nick_name"`
	Type     PlayerType `json:"type"`
	IsOnline bool       `json:"is_online"`
	JoinTime time.Time  `json:"join_time"`
}
