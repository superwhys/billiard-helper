package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/infra/socket"
)

type eventTestHook struct{ socket.SessionHook }

func (eventTestHook) OnAllowRequest(r *http.Request) (*http.Request, error) { return r, nil }

type eventTestConn struct {
	websocketutils.Conn
	id     string
	closed bool
}

func (c *eventTestConn) ID() string { return c.id }
func (c *eventTestConn) Close() error {
	c.closed = true
	return nil
}

func TestUserBlockedEventClosesConnection(t *testing.T) {
	sm := socket.NewSocketManager(eventTestHook{})
	conn := &eventTestConn{id: "target"}
	sm.RegisterSession(7, conn)
	h := NewHandlers(sm, nil)

	h.Call(context.Background(), constant.EventUserBlocked, []byte(`{"user_id":7}`))
	if !conn.closed {
		t.Fatal("user blocked event did not close the target connection")
	}
}
