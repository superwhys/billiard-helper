package service

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/ports"
)

type scoresService struct {
	srvCtx *ServiceContext
}

func NewScoresService(srvCtx *ServiceContext) ports.ScoresService {
	return &scoresService{
		srvCtx: srvCtx,
	}
}

func (s *scoresService) AddScore(ctx context.Context, req *request.AddScoreRequest) error {
	panic("not implemented")
}

func (s *scoresService) MinusScore(ctx context.Context, req *request.MinusScoreRequest) error {
	panic("not implemented")
}

func (s *scoresService) ResetScore(ctx context.Context, req *request.ResetScoreRequest) error {
	panic("not implemented")
}

func (s *scoresService) GetRoomScores(ctx context.Context, req *request.GetRoomScoresRequest) ([]*response.Scores, error) {
	panic("not implemented")
}
