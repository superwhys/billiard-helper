package dto

type Operator struct {
	UserID    uint   `json:"-"`
	SessionID string `json:"-"`
}
