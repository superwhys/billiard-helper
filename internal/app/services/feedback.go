package services

import (
	"context"
	"strings"

	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/domain/feedback"
	"github.com/superwhys/billiard-helper/internal/errcode"
)

type FeedbackApp struct {
	repoFactory factory.IRepoFactory
}

func NewFeedbackApp(repoFactory factory.IRepoFactory) *FeedbackApp {
	return &FeedbackApp{
		repoFactory: repoFactory,
	}
}

func (a *FeedbackApp) ReportFeedback(ctx context.Context, req *dto.FeedbackReportReq) error {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return errcode.ErrBadRequest.WithMessage("反馈内容不能为空")
	}

	feedbackRepo := a.repoFactory.FeedbackRepo()
	return feedbackRepo.Create(ctx, &feedback.Feedback{
		UserID:  req.UserID,
		Content: content,
	})
}
