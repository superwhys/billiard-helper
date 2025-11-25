package ports

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/models/dbmodels"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user *dbmodels.User) error
	GetUserByEmail(ctx context.Context, email string) (*dbmodels.User, error)
	GetUserByID(ctx context.Context, id uint) (*dbmodels.User, error)
	UpdateUser(ctx context.Context, user *dbmodels.User) error
	DeleteUser(ctx context.Context, id uint) error
}

type RoomRepo interface {
	CreateRoom(ctx context.Context, room *dbmodels.Room) error
	IsRoomExist(ctx context.Context, roomID uint) (bool, error)
	GetRoom(ctx context.Context, roomID uint) (*dbmodels.Room, error)
	GetUserRooms(ctx context.Context, userID uint) ([]*dbmodels.Room, error)
	DeleteRoom(ctx context.Context, roomID uint) error
}

type PlayerRepo interface {
	CreatePlayer(ctx context.Context, player *dbmodels.Player) error
	GetPlayer(ctx context.Context, playerCode string) (*dbmodels.Player, error)
	GetPlayerByCode(ctx context.Context, code string) (*dbmodels.Player, error)
	UpdatePlayer(ctx context.Context, playerCode string, player *dbmodels.Player) error
	DeletePlayer(ctx context.Context, playerCode string) error
}

type ScoreRepo interface {
	CreateScore(ctx context.Context, score *dbmodels.Scores) error
	DeleteScore(ctx context.Context, scoreID uint) error
}
