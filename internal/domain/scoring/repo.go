package scoring

import "context"

// IScoringRepository 仓储接口
type IScoringRepository interface {
	// AddEvent 记录一次分数变化
	AddEvent(ctx context.Context, event *Score) error
	// GetRoomEvents 获取房间的所有事件（用于重放计算）
	GetRoomEvents(ctx context.Context, roomID uint) ([]*Score, error)
	// DeleteLastEvent 删除(或标记失效)最后一条事件 -> 对应撤回
	DeleteLastEvent(ctx context.Context, roomID uint) error
}
