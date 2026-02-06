package event

import "context"

// IEventService 领域服务接口
type IEventService interface {
	// RecordScore 记录分数
	AddEvent(ctx context.Context, event *Event) error
	// Undo 撤回最近一次操作
	Undo(ctx context.Context, matchID uint) error
}
