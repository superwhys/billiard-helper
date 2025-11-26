package socket

import (
	"net/http"

	"github.com/miebyte/goutils/websocketutils"
)

// SessionHook 定义了外部需要实现的回调
// Comet 只知道需要验证 Token，不知道具体怎么验证
type SessionHook interface {
	OnConnect(ctx *websocketutils.Context)
	OnDisconnect(ctx *websocketutils.Context)
	OnAllowRequest(request *http.Request) (*http.Request, error)
}
