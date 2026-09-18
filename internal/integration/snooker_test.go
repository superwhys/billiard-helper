package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/redisutils"
	"github.com/redis/go-redis/v9"
	"github.com/superwhys/billiard-helper/api/routers"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	appfactory "github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/infra/cache"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"github.com/superwhys/billiard-helper/internal/infra/eventbus"
	infrafactory "github.com/superwhys/billiard-helper/internal/infra/factory"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Requires a dedicated empty test database and Redis instance; no production
// configuration is read. The caller owns disposal of these test services.
func TestSnookerHTTPPersistence(t *testing.T) {
	dsn, redisAddr := os.Getenv("BILLIARD_TEST_MYSQL_DSN"), os.Getenv("BILLIARD_TEST_REDIS_ADDR")
	if dsn == "" || redisAddr == "" {
		t.Skip("set BILLIARD_TEST_MYSQL_DSN and BILLIARD_TEST_REDIS_ADDR for isolated integration tests")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := db.AutoMigrate(models.Tables()...); err != nil {
		t.Fatal(err)
	}
	owner, stranger := &models.User{Name: "snooker test owner"}, &models.User{Name: "snooker test stranger"}
	if err := db.Create(owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(stranger).Error; err != nil {
		t.Fatal(err)
	}
	rdb := &redisutils.RedisClient{Client: redis.NewClient(&redis.Options{Addr: redisAddr})}
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatal(err)
	}
	repos := infrafactory.NewRepositoryFactory(db)
	domain := appfactory.NewDomainServiceFactory(rdb, nil)
	bus := eventbus.NewRedisEventBus(rdb)
	matches := services.NewMatchApp(domain, repos, bus, cache.NewLockManager(rdb))
	scores := services.NewScoreApp(domain, repos, bus)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		userID := owner.ID
		if c.GetHeader("X-Test-User") == "other" {
			userID = stranger.ID
		}
		c.Request = c.Request.WithContext(jwt.SetTokenClaimsToContext(c.Request.Context(), &jwt.UserTokenClaims{UserID: userID}))
	})
	router.POST("/match/create", routers.MatchCreateHandler(matches))
	router.POST("/match/start", routers.MatchStartHandler(matches))
	router.POST("/match/update", routers.MatchUpdateHandler(matches))
	router.POST("/match/round/next", routers.MatchRoundNextHandler(matches))
	router.POST("/score/sync", routers.ScoreSyncHandler(scores))
	router.POST("/score/undo", routers.ScoreUndoHandler(scores))
	router.GET("/match/detail", routers.MatchDetailHandler(matches))
	router.GET("/match/list", routers.MatchListHandler(matches))
	call := func(method, path string, data any, other bool) response {
		body, _ := json.Marshal(data)
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if other {
			req.Header.Set("X-Test-User", "other")
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var res response
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Errorf("invalid response %s: %v", rec.Body.String(), err)
		}
		return res
	}
	requireOK := func(res response) {
		t.Helper()
		if res.Code != 0 {
			t.Fatalf("request failed: %d %s", res.Code, res.Message)
		}
	}
	requireFail := func(res response, fragment string) {
		t.Helper()
		if res.Code == 0 || !strings.Contains(res.Message, fragment) {
			t.Fatalf("wanted %q error: %+v", fragment, res)
		}
	}
	creation := map[string]any{"name": "integration snooker", "match_type": "snooker", "target_score": 3, "virtual_players": []map[string]any{{"nick_name": "A", "type": 1}, {"nick_name": "B", "type": 1}}}
	res := call(http.MethodPost, "/match/create", creation, false)
	requireOK(res)
	var m dto.Match
	if err := json.Unmarshal(res.Data, &m); err != nil {
		t.Fatal(err)
	}
	if m.MatchRound != 1 || len(m.Players) != 2 {
		t.Fatalf("bad creation %+v", m)
	}
	first, second := m.Players[0].ID, m.Players[1].ID
	detail := func() dto.Match {
		t.Helper()
		res := call(http.MethodGet, fmt.Sprintf("/match/detail?match_id=%d", m.ID), nil, false)
		requireOK(res)
		var result dto.Match
		if err := json.Unmarshal(res.Data, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	action := map[string]any{"match_id": m.ID}
	requireFail(call(http.MethodPost, "/match/start", action, true), "比赛不存在")
	requireOK(call(http.MethodPost, "/match/start", action, false))
	event := func(key string, scorer, recipient uint, points int) map[string]any {
		return map[string]any{"match_id": m.ID, "round": 1, "score_actions": []map[string]any{{"player_ids": []uint{recipient}, "score": points}}, "context": map[string]any{"stat_key": key, "scorer_player_id": scorer}}
	}
	requireFail(call(http.MethodPost, "/score/sync", event("red", first, first, 1), true), "比赛不存在")
	requireFail(call(http.MethodPost, "/score/sync", event("red", first, first, 100), false), "斯诺克计分无效")
	var eventCount int64
	db.Model(&models.MatchEvent{}).Where("match_id = ?", m.ID).Count(&eventCount)
	if eventCount != 0 {
		t.Fatal("invalid score event was not rolled back")
	}
	requireOK(call(http.MethodPost, "/score/sync", event("black", first, first, 7), false))
	requireOK(call(http.MethodPost, "/score/sync", event("foul", second, first, 4), false))
	requireOK(call(http.MethodPost, "/score/undo", map[string]any{"match_id": m.ID, "round": 1}, false))
	d := detail()
	var snapshot map[string]struct {
		Score int `json:"score"`
	}
	raw, _ := json.Marshal(d.CurrentScores)
	json.Unmarshal(raw, &snapshot)
	if snapshot[fmt.Sprint(first)].Score != 7 || snapshot[fmt.Sprint(second)].Score != 0 {
		t.Fatalf("undo not persisted %s", raw)
	}
	// Concurrent writes must reload the snapshot after acquiring the row lock.
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := call(http.MethodPost, "/score/sync", event("red", first, first, 1), false)
			if res.Code != 0 {
				t.Errorf("concurrent score: %s", res.Message)
			}
		}()
	}
	wg.Wait()
	d = detail()
	raw, _ = json.Marshal(d.CurrentScores)
	json.Unmarshal(raw, &snapshot)
	if snapshot[fmt.Sprint(first)].Score != 19 {
		t.Fatalf("concurrent points lost: %s", raw)
	}
	settle := map[string]any{"match_id": m.ID, "round": 1}
	requireFail(call(http.MethodPost, "/match/round/next", settle, true), "比赛不存在")
	requireOK(call(http.MethodPost, "/match/round/next", settle, false))
	requireFail(call(http.MethodPost, "/match/round/next", settle, false), "比赛轮数不匹配")
	requireFail(call(http.MethodPost, "/score/sync", event("red", first, first, 1), false), "比赛轮数不匹配")
	requireFail(call(http.MethodPost, "/score/undo", map[string]any{"match_id": m.ID, "round": 1}, false), "比赛轮数不匹配")
	d = detail()
	if d.MatchRound != 2 || len(d.MatchGames) != 2 || d.Players[0].Scores != 1 {
		t.Fatalf("frame not saved %+v", d)
	}
	requireFail(call(http.MethodPost, "/match/round/next", map[string]any{"match_id": m.ID, "round": 2}, false), "本局比分相同")
	requireFail(call(http.MethodPost, "/match/round/next", map[string]any{"match_id": m.ID, "round": 2, "conceding_player_id": 999999}, false), "认输球员不属于")
	e := event("blue", second, second, 5)
	e["round"] = 2
	requireOK(call(http.MethodPost, "/score/sync", e, false))
	requireOK(call(http.MethodPost, "/match/round/next", map[string]any{"match_id": m.ID, "round": 2, "conceding_player_id": second}, false))
	d = detail()
	if d.Status != 3 || d.WinnerID == nil || *d.WinnerID != first || d.WinnerScore != 2 || len(d.MatchGames) != 2 {
		t.Fatalf("result not persisted %+v", d)
	}
	if d.MatchGames[0].WinnerID == nil || *d.MatchGames[0].WinnerID != first {
		t.Fatal("concession winner not persisted")
	}
	var frame map[string]struct {
		Score int `json:"score"`
	}
	json.Unmarshal(d.MatchGames[0].Scores, &frame)
	if frame[fmt.Sprint(second)].Score != 5 {
		t.Fatal("concession modified actual score")
	}
	requireFail(call(http.MethodPost, "/score/sync", e, false), "比赛未开始")
	res = call(http.MethodGet, "/match/list?match_type=snooker&limit=10", nil, false)
	requireOK(res)
	var list []dto.Match
	json.Unmarshal(res.Data, &list)
	if len(list) == 0 || list[0].ID != m.ID || list[0].WinnerScore != 2 {
		t.Fatalf("list result incorrect: %s", res.Data)
	}
	// Invalid setup must fail before creating rows.
	creation["target_score"] = 2
	requireFail(call(http.MethodPost, "/match/create", creation, false), "奇数局")
}
