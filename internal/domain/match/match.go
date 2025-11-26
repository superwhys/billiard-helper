package match

import "time"

// Room 聚合根
type Room struct {
	ID        uint
	RoomCode  string
	OwnerID   uint
	Status    RoomStatus
	Config    *GameConfig
	Players   []*Player
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Player struct {
	ID       uint
	RoomID   uint
	UserID   *uint
	NickName string
	Type     PlayerType
	IsOnline bool
	JoinTime time.Time
}
