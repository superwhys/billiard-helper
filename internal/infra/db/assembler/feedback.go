package assembler

import (
	"github.com/superwhys/billiard-helper/internal/domain/feedback"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/gorm"
)

type FeedbackPoAssembler struct{}

func NewFeedbackPoAssembler() *FeedbackPoAssembler {
	return &FeedbackPoAssembler{}
}

func (a *FeedbackPoAssembler) ToEntity(po *models.Feedback) *feedback.Feedback {
	if po == nil {
		return nil
	}

	return &feedback.Feedback{
		ID:      po.ID,
		UserID:  po.UserID,
		Content: po.Content,
	}
}

func (a *FeedbackPoAssembler) ToPO(entity *feedback.Feedback) *models.Feedback {
	if entity == nil {
		return nil
	}

	return &models.Feedback{
		Model: gorm.Model{
			ID: entity.ID,
		},
		UserID:  entity.UserID,
		Content: entity.Content,
	}
}
