package service

import "github.com/superwhys/billiard-helper/ports"

type Service struct {
	SocketService ports.SocketService
}

func NewService() *Service {
	return &Service{}
}
