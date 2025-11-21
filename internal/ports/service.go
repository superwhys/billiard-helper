package ports

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

// TokenAuthLogic Token 认证用例接口
type TokenAuthLogic interface {
	GetUserTokenClaims(context.Context, string) (*jwt.UserTokenClaims, error)
}

type AuthService interface {
	TokenAuthLogic
	SendEmailCode(context.Context, *request.SendEmailCodeReq) error
	Register(context.Context, *request.RegisterReq) error
	Login(context.Context, *request.LoginReq) (string, error)
}

type RoomService interface {
	CreateRoom(ctx context.Context, req *request.CreateRoomRequest) (*response.Room, error)
	GetRoom(ctx context.Context, req *request.GetRoomRequest) (*response.Room, error)
	GetUserRooms(ctx context.Context, req *request.GetUserRoomsRequest) ([]*response.Room, error)
	JoinRoom(ctx context.Context, req *request.JoinRoomRequest) (*response.Room, error)
	LeaveRoom(ctx context.Context, req *request.LeaveRoomRequest) error
	DeleteRoom(ctx context.Context, req *request.DeleteRoomRequest) error
}

type ScoresService interface {
	AddScore(ctx context.Context, req *request.AddScoreRequest) error
	MinusScore(ctx context.Context, req *request.MinusScoreRequest) error
	ResetScore(ctx context.Context, req *request.ResetScoreRequest) error
	GetRoomScores(ctx context.Context, req *request.GetRoomScoresRequest) ([]*response.Scores, error)
}

type App interface {
	RoomService
	ScoresService
}
