package feedback

import "context"

type IFeedbackRepository interface {
	Create(ctx context.Context, feedback *Feedback) error
}
