package services

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/app/assembler"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
)

type ScoreApp struct {
	serviceFactory *factory.DomainServiceFactory
	repoFactory    factory.IRepoFactory
	scoreAssembler *assembler.ScoreAssembler
	eventBus       shared.EventBus
}

func NewScoreApp(
	serviceFactory *factory.DomainServiceFactory,
	repoFactory factory.IRepoFactory,
	eventBus shared.EventBus,
) *ScoreApp {
	return &ScoreApp{
		serviceFactory: serviceFactory,
		repoFactory:    repoFactory,
		scoreAssembler: assembler.NewScoreAssembler(),
		eventBus:       eventBus,
	}
}

// AddScore 记录加分/减分/犯规
func (a *ScoreApp) AddScore(ctx context.Context, req *dto.AddScoreRequest) (*dto.RoomScoreResponse, error) {
	panic("not implemented")
}

// Undo 撤回
func (a *ScoreApp) Undo(ctx context.Context, req *dto.UndoRequest) (*dto.RoomScoreResponse, error) {
	panic("not implemented")
}
