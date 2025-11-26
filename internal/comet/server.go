package comet

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/sourcegraph/conc/pool"
	"github.com/superwhys/billiard-helper/internal/comet/consume"
	"github.com/superwhys/billiard-helper/internal/comet/manager"
	"github.com/superwhys/billiard-helper/internal/comet/queue"
	"github.com/superwhys/billiard-helper/internal/models/constant"
	"github.com/superwhys/billiard-helper/internal/service"
)

type Server struct {
	srv            *service.Service
	sessionManager *manager.SessionManager
	handlers       *consume.Handlers
	queue          queue.EventQueue
}

func NewCometServer(queue queue.EventQueue, srv *service.Service) *Server {
	sessionManager := manager.NewSessionManager(srv)
	handlers := consume.NewHandlers(srv, sessionManager)

	server := &Server{
		srv:            srv,
		sessionManager: sessionManager,
		handlers:       handlers,
		queue:          queue,
	}

	return server
}

func (s *Server) Subscribe(ctx context.Context) error {
	ch := s.queue.Subscribe(ctx, constant.BilliardMessageChannel)

	worker := pool.New().WithMaxGoroutines(10)
	defer worker.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil
		case data := <-ch:
			worker.Go(func() {
				s.call(ctx, data)
			})
		}
	}
}

func (s *Server) call(ctx context.Context, data []byte) {
	var msg queue.QueueMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	logging.Debugc(ctx, "received event(%s): %s", msg.Event, string(msg.Data))

	s.handlers.Call(ctx, msg.Event, msg.Data)
}

func (s *Server) Handler() ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/ws"),
		ginutils.WithHandler(http.MethodGet, "", s.sessionManager.ServeHttp()),
		ginutils.WithHandler(http.MethodGet, "/:path", s.sessionManager.ServeHttp()),
	)
}
