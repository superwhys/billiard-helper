package db

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/feedback"
	"github.com/superwhys/billiard-helper/internal/infra/db/assembler"
	"github.com/superwhys/billiard-helper/internal/infra/db/query"
	"gorm.io/gorm"
)

var _ feedback.IFeedbackRepository = (*FeedbackRepo)(nil)

type FeedbackRepo struct {
	query               *query.Query
	feedbackPoAssembler *assembler.FeedbackPoAssembler
}

func NewFeedbackRepo(db *gorm.DB) *FeedbackRepo {
	return &FeedbackRepo{
		query:               query.Use(db),
		feedbackPoAssembler: assembler.NewFeedbackPoAssembler(),
	}
}

func (r *FeedbackRepo) Create(ctx context.Context, feedbackEntity *feedback.Feedback) error {
	po := r.feedbackPoAssembler.ToPO(feedbackEntity)
	err := r.query.Feedback.WithContext(ctx).Create(po)
	if err != nil {
		return err
	}

	feedbackEntity.ID = po.ID
	return nil
}
