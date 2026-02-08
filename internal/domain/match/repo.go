package match

import "context"

type IMatchRepository interface {
	Create(ctx context.Context, match *Match) error
	FindByID(ctx context.Context, id uint) (*Match, error)
	IsExists(ctx context.Context, id uint) (bool, error)
	Update(ctx context.Context, match *Match) error
	ListMatches(ctx context.Context, matchType string, limit uint, cursor uint) ([]*Match, error)
	Delete(ctx context.Context, id uint) error
}

type IPlayerRepository interface {
	FindByCode(ctx context.Context, code string) (*Player, error)
	Create(ctx context.Context, player *Player) error
	CreateInBatches(ctx context.Context, players []*Player) error
	Delete(ctx context.Context, id uint) error
}
