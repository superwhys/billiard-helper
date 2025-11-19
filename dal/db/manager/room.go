package manager

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/superwhys/billiard-helper/dal/db/query"
	"github.com/superwhys/billiard-helper/models/dbmodels"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/request"
	"github.com/superwhys/billiard-helper/models/types"
	"github.com/superwhys/billiard-helper/ports"
)

type roomManager struct {
	query *query.Query
}

// NewRoomManager 创建房间仓储
func NewRoomManager(db *gorm.DB) ports.RoomRepo {
	return &roomManager{
		query: query.Use(db),
	}
}

func (m *roomManager) CreateRoom(ctx context.Context, room *request.CreateRoomRequest) error {
	if room == nil {
		return errcode.ErrCodeInvalidRequest
	}

	roomModel := &dbmodels.Room{
		RoomCode: room.RoomCode,
		UserID:   room.UserID,
		Status:   types.RoomStatusPending,
	}
	return m.query.Room.WithContext(ctx).Create(roomModel)
}

func (m *roomManager) GetRoom(ctx context.Context, roomID uint) (*dbmodels.Room, error) {
	r := m.query.Room
	room, err := r.WithContext(ctx).
		Where(r.ID.Eq(roomID)).
		Preload(r.Players.RelationField).
		Preload(r.Scores.RelationField).
		First()

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if room == nil {
		return nil, errcode.ErrCodeRoomNotFound
	}

	return room, nil
}

func (m *roomManager) GetUserRooms(ctx context.Context, userID uint) ([]*dbmodels.Room, error) {
	r := m.query.Room
	return r.WithContext(ctx).
		Where(r.UserID.Eq(userID)).
		Preload(r.Players.RelationField).
		Preload(r.Scores.RelationField).
		Find()
}

func (m *roomManager) DeleteRoom(ctx context.Context, roomID uint) error {
	room := m.query.Room
	_, err := room.WithContext(ctx).
		Where(room.ID.Eq(roomID)).
		Delete(&dbmodels.Room{})
	return err
}
