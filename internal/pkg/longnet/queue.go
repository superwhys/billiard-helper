package longnet

import (
	"sync"
)

type MemoryQueue struct {
	mu       sync.RWMutex
	handlers map[string][]func(data []byte)
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		handlers: make(map[string][]func(data []byte)),
	}
}

func (m *MemoryQueue) Subscribe(channel string, callback func(data []byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[channel] = append(m.handlers[channel], callback)
}

func (m *MemoryQueue) Publish(channel string, data []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if handlers, ok := m.handlers[channel]; ok {
		for _, handler := range handlers {
			go handler(data)
		}
	}
}
