package assembler

import (
	"encoding/json"

	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/domain/scoring"
)

type ScoreAssembler struct{}

func NewScoreAssembler() *ScoreAssembler {
	return &ScoreAssembler{}
}

// ToScoreEventDTO 将领域对象转换为 DTO
func (a *ScoreAssembler) ToScoreDTO(e *scoring.Score) dto.Score {
	if e == nil {
		return dto.Score{}
	}
	return dto.Score{
		ID:         e.ID,
		RoomID:     e.RoomID,
		PlayerID:   e.PlayerID,
		OperatorID: e.OperatorID,
		Type:       e.Type,
		Value:      e.Value,
		Context:    e.Context,
	}
}

// ToScoreEventEntity 将请求 DTO 转换为领域实体
// 注意：需要外部提供 RoomID，因为 DTO 只包含 RoomCode
func (a *ScoreAssembler) ToScoreEntity(req *dto.AddScoreRequest, roomID uint) (*scoring.Score, error) {
	ctxBytes, err := json.Marshal(req.Context)
	if err != nil {
		return nil, err
	}

	return &scoring.Score{
		RoomID:     roomID,
		PlayerID:   req.PlayerID,
		OperatorID: req.OperatorID,
		Type:       req.Type,
		Value:      req.Value,
		Context:    string(ctxBytes),
	}, nil
}
