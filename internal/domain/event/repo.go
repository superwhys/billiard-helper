package event

import "context"

// IEventRepository 仓储接口
type IEventRepository interface {
	// AddEvent 添加事件
	AddEvent(ctx context.Context, event *Event) error
	// GetMatchEvents 获取比赛的所有事件（用于重放计算）
	GetMatchEvents(ctx context.Context, matchID uint) ([]*Event, error)
	// DeleteLastEvent 删除(或标记失效)最后一条事件 -> 对应撤回操作
	DeleteLastEvent(ctx context.Context, matchID uint) error
}
