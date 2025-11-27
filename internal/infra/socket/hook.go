package socket

import (
	"context"
	"net/http"

	"github.com/miebyte/goutils/websocketutils"
)

// SessionHook 定义了外部需要实现的回调
// Comet 只知道需要验证 Token，不知道具体怎么验证
type SessionHook interface {
	OnConnect(ctx context.Context) (uint, error)
	OnDisconnect(ctx context.Context, conn websocketutils.Conn)
	OnAllowRequest(request *http.Request) (*http.Request, error)
}
