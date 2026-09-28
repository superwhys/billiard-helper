package hook

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/infra/socket"
	"github.com/superwhys/billiard-helper/internal/pkg/codegen"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

var _ socket.SessionHook = (*SocketHook)(nil)

type SocketHook struct {
	matchApp *services.MatchApp
	userApp  *services.UserApp
}

func NewSocketHook(matchApp *services.MatchApp, userApp *services.UserApp) *SocketHook {
	return &SocketHook{
		matchApp: matchApp,
		userApp:  userApp,
	}
}

func (h *SocketHook) OnConnect(ctx context.Context) (uint, error) {
	claims, err := jwt.TokenClaimsFromContext(ctx)
	if err != nil {
		logging.Errorc(ctx, "get token claims from context failed: %v", err)
		return 0, err
	}
	blocked, err := h.userApp.IsUserBlocked(ctx, claims.UserID)
	if err != nil {
		return 0, err
	}
	if blocked {
		return 0, errcode.ErrForbidden
	}

	return claims.UserID, nil
}

func (h *SocketHook) OnDisconnect(ctx context.Context, conn websocketutils.Conn) {
	claims, err := jwt.TokenClaimsFromContext(ctx)
	if err != nil {
		logging.Errorc(ctx, "get token claims from context failed: %v", err)
		return
	}

	logging.Infoc(ctx, "user(%d) disconnect", claims.UserID)

	// 通知所有加入的房间。
	enterRooms := conn.Rooms()
	for _, room := range enterRooms {
		if !constant.IsSocketRoomID(room) {
			continue
		}
		logging.Infoc(ctx, "user(%d) leave room(%s)", claims.UserID, room)
		matchID := constant.ParseSocketRoomID(room)
		if matchID == 0 {
			continue
		}

		playerCode := codegen.GeneratePlayerCode(matchID, uint8(match.PlayerTypeReal), fmt.Sprintf("%d", claims.UserID))

		err = h.matchApp.LeaveMatch(ctx, &dto.MatchActionRequest{
			Operator:   dto.Operator{UserID: claims.UserID},
			MatchID:    constant.ParseSocketRoomID(room),
			PlayerCode: playerCode,
		})

		if err != nil {
			logging.Errorc(ctx, "leave match failed: %v", err)
		}
	}
}

func (h *SocketHook) OnAllowRequest(r *http.Request) (*http.Request, error) {
	ctx := logging.CloneContext(r.Context())
	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" {
		return nil, errcode.ErrUnauthorized
	}

	// Support "Bearer <token>" format
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	claims, err := h.userApp.GetUserTokenClaims(ctx, tokenStr)
	if err != nil {
		logging.Errorc(ctx, "verify websocket token failed: %v", err)
		if ec, ok := errcode.AsErrcode(err); ok {
			return nil, ec
		}
		return nil, errcode.ErrUnauthorized
	}

	reqCtx := jwt.SetTokenClaimsToContext(r.Context(), claims)
	return r.WithContext(reqCtx), nil
}

func (h *SocketHook) IsUserBlocked(ctx context.Context, userID uint) (bool, error) {
	return h.userApp.IsUserBlocked(ctx, userID)
}
