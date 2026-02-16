package factory

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/event"
	"github.com/superwhys/billiard-helper/internal/domain/feedback"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/user"
)

type IRepoFactory interface {
	UserRepo() user.IUserRepository
	MatchRepo() match.IMatchRepository
	PlayerRepo() match.IPlayerRepository
	MatchGameRepo() match.IMatchGameRepository
	EventRepo() event.IEventRepository
	FeedbackRepo() feedback.IFeedbackRepository
	WithTransaction(ctx context.Context, fn func(factory IRepoFactory) error) error
}
