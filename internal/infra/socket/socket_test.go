package socket

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/miebyte/goutils/websocketutils"
)

type testSessionHook struct {
	blocked map[uint]bool
	err     error
}

func (h *testSessionHook) OnConnect(context.Context) (uint, error)           { return 0, nil }
func (h *testSessionHook) OnDisconnect(context.Context, websocketutils.Conn) {}
func (h *testSessionHook) OnAllowRequest(r *http.Request) (*http.Request, error) {
	return r, nil
}
func (h *testSessionHook) IsUserBlocked(_ context.Context, userID uint) (bool, error) {
	return h.blocked[userID], h.err
}

type testConn struct {
	websocketutils.Conn
	id      string
	closed  int
	emitted int
	onClose func()
}

func (c *testConn) ID() string { return c.id }
func (c *testConn) Close() error {
	c.closed++
	if c.onClose != nil {
		c.onClose()
	}
	return nil
}
func (c *testConn) Emit(string, any) error {
	c.emitted++
	return nil
}

func TestCloseUserConnectionsClosesEveryConnection(t *testing.T) {
	sm := NewSocketManager(&testSessionHook{})
	first := &testConn{id: "first"}
	second := &testConn{id: "second"}
	other := &testConn{id: "other"}
	for _, conn := range []*testConn{first, second} {
		sm.RegisterSession(1, conn)
		conn.onClose = func() { sm.UnregisterSession(conn) }
	}
	sm.RegisterSession(2, other)

	sm.CloseUserConnections(1)
	if first.closed != 1 || second.closed != 1 || other.closed != 0 {
		t.Fatalf("closed counts: first=%d second=%d other=%d", first.closed, second.closed, other.closed)
	}
	if len(sm.userSessions(1)) != 0 || len(sm.userSessions(2)) != 1 {
		t.Fatal("closing one user removed the wrong connections")
	}
}

func TestBroadcastToRoomSkipsBlockedConnections(t *testing.T) {
	hook := &testSessionHook{blocked: map[uint]bool{1: true}}
	sm := NewSocketManager(hook)
	room := sm.billiardNamespace.Room("match_1")
	blocked := &testConn{id: "blocked"}
	allowed := &testConn{id: "allowed"}
	for userID, conn := range map[uint]*testConn{1: blocked, 2: allowed} {
		sm.RegisterSession(userID, conn)
		room.Add(conn)
		conn.onClose = func() {
			room.Remove(conn)
			sm.UnregisterSession(conn)
		}
	}

	if err := sm.BroadcastToRoom(context.Background(), "match_1", "score", 1); err != nil {
		t.Fatal(err)
	}
	if blocked.closed != 1 || blocked.emitted != 0 || allowed.closed != 0 || allowed.emitted != 1 {
		t.Fatalf("blocked close/emit=%d/%d, allowed close/emit=%d/%d", blocked.closed, blocked.emitted, allowed.closed, allowed.emitted)
	}

	hook.err = errors.New("database unavailable")
	if err := sm.BroadcastToRoom(context.Background(), "match_1", "score", 2); err == nil {
		t.Fatal("expected status-check error")
	}
	if allowed.closed != 1 || allowed.emitted != 1 {
		t.Fatal("status-check failure must stop sending")
	}
}

func TestGetUserSessionsRejectsBlockedUser(t *testing.T) {
	sm := NewSocketManager(&testSessionHook{blocked: map[uint]bool{1: true}})
	conn := &testConn{id: "blocked"}
	sm.RegisterSession(1, conn)
	conn.onClose = func() { sm.UnregisterSession(conn) }

	if sessions := sm.GetUserSessions(context.Background(), 1); len(sessions) != 0 {
		t.Fatal("blocked user sessions were returned")
	}
	if conn.closed != 1 {
		t.Fatal("blocked user connection was not closed")
	}
}
