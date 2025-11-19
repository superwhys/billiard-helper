package types

type Player struct {
	Code      string     `json:"code"`
	RoomID    uint       `json:"room_id"`
	UserID    uint       `json:"user_id"`
	NickName  string     `json:"nick_name"`
	AvatarURL string     `json:"avatar_url"`
	Type      PlayerType `json:"type"`

	Scores []*Scores `json:"scores,omitempty"`
}
