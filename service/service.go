package service

import "github.com/superwhys/billiard-helper/ports"

type Service struct {
	RoomService ports.RoomService
}

func NewService() *Service {
	return &Service{}
}
