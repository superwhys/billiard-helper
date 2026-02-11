package event

import "context"

// IEventRepository 仓储接口
type IEventRepository interface {
	// AddEvent 添加事件
	AddEvent(ctx context.Context, event *Event) error
	// GetMatchEvents 获取比赛的所有事件（用于重放计算）
	GetMatchEvents(ctx context.Context, matchID uint) ([]*Event, error)
	// DeleteEvent 删除(或标记失效)指定轮次的事件 -> 对应撤回操作
	DeleteEvent(ctx context.Context, eventID uint) (*Event, error)
	GetLastEvent(ctx context.Context, matchID uint, round uint) (*Event, error)
}
