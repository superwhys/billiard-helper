package types

type Player struct {
	ID        uint       `json:"id"`
	Code      string     `json:"code"`
	RoomID    uint       `json:"room_id"`
	UserID    uint       `json:"user_id"`
	NickName  string     `json:"nick_name"`
	AvatarURL string     `json:"avatar_url"`
	Type      PlayerType `json:"type"`
	IsOnline  bool       `json:"is_online"`

	// 用来标识返回给客户端的数据时，是否是当前客户端用户
	IsYou bool `json:"is_you,omitempty"`

	Scores []*Scores `json:"scores,omitempty"`
}
