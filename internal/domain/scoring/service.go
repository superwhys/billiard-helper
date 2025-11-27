package scoring

import "context"

// IScoringService 领域服务接口
type IScoringService interface {
	// RecordScore 记录分数
	RecordScore(ctx context.Context, event *Score) error
	// Undo 撤回最近一次操作
	Undo(ctx context.Context, roomID uint, operatorID uint) error
}
