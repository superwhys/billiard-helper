package types

type User struct {
	ID     uint   `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`

	Rooms []*Room `json:"rooms,omitempty"`
}
