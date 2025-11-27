package match

import "context"

type IMatchRepository interface {
	Create(ctx context.Context, room *Room) error
	FindByID(ctx context.Context, id uint) (*Room, error)
	FindByCode(ctx context.Context, code string) (*Room, error)
	Update(ctx context.Context, room *Room) error
	Delete(ctx context.Context, id uint) error
}
