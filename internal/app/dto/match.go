package dto

import (
	"time"
)

type GameConfig struct {
	GameType    int `json:"game_type"` // 比如 1: snooker, 2: 8-ball
	MaxPlayers  int `json:"max_players"`
	TargetScore int `json:"target_score"`
}

type Room struct {
	ID        uint       `json:"id"`
	RoomCode  string     `json:"room_code"`
	OwnerID   uint       `json:"owner_id"`
	Status    int        `json:"status"`
	Config    GameConfig `json:"config"`
	Players   []Player   `json:"players"`
	CreatedAt time.Time  `json:"created_at"`
}

type Player struct {
	ID       uint      `json:"id"`
	UserID   *uint     `json:"user_id,omitempty"`
	NickName string    `json:"nick_name"`
	Type     int       `json:"type"`
	IsOnline bool      `json:"is_online"`
	JoinTime time.Time `json:"join_time"`
}

type Operator struct {
	UserID uint `json:"-"`
}

type CreateRoomRequest struct {
	Operator
	GameType    int `json:"game_type"`
	MaxPlayers  int `json:"max_players"`
	TargetScore int `json:"target_score"`
}

type JoinRoomRequest struct {
	Operator
	RoomID   uint   `json:"room_id"`
	NickName string `json:"nick_name"`
}

type RoomActionRequest struct {
	Operator
	RoomID uint `json:"room_id"`
}

type KickPlayerRequest struct {
	Operator
	RoomID       uint `json:"room_id"`
	TargetUserID uint `json:"target_user_id"`
}
