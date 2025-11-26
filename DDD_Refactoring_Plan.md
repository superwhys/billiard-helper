# DDD 架构重构方案

根据当前项目需求与业务逻辑，以下是将 Billiard Helper 项目重构为 DDD（领域驱动设计）架构的详细方案。

## 1. 领域层划分 (Domain Layer Division)

我们将核心业务划分为三个主要的领域（Bounded Contexts）：

1.  **用户与身份领域 (Identity / User Context)**
    -   负责用户的注册、登录认证、个人信息管理。
    -   作为基础支撑领域，为其他领域提供用户身份标识。

2.  **比赛/房间领域 (Match / Room Context)**
    -   核心业务领域。
    -   负责房间的生命周期管理（创建、开启、结束）。
    -   负责房间内的参与者（Player）管理（加入、离开、踢出）。
    -   负责配置比赛类型（如中式八球、斯诺克），但仅作为元数据存储。

3.  **记分领域 (Scoring Context)**
    -   核心业务领域。
    -   负责比赛过程中的通用得分记录（加分/减分）。
    -   负责撤回操作（Undo）及分数统计。
    -   采用类事件溯源（Event Sourcing）的方式记录分数变化。
    -   **核心变更**：不包含复杂规则计算，分值变化由客户端传入，后端仅负责记录和统计。

---

## 2. 领域定义 (Domain Definitions)

### 2.1 用户领域 (User Domain)

**Package**: `internal/domain/user`

#### 值对象 (Value Objects)

```go
package user

import (
	"errors"
	"regexp"
	"golang.org/x/crypto/bcrypt"
)

// Email 值对象
type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	if email == "" {
		return Email{}, errors.New("email cannot be empty")
	}
	// 简单的正则校验
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	if !match {
		return Email{}, errors.New("invalid email format")
	}
	return Email{value: email}, nil
}

func (e Email) String() string { return e.value }

// Password 值对象
type Password struct {
	hash string
}

// NewPasswordFromPlain 从明文创建
func NewPasswordFromPlain(plain string) (Password, error) {
	if len(plain) < 6 {
		return Password{}, errors.New("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, err
	}
	return Password{hash: string(hash)}, nil
}

// NewPasswordFromHash 从哈希加载
func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

func (p Password) Hash() string { return p.hash }

func (p Password) Compare(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plain))
	return err == nil
}
```

#### 聚合根与实体 (Aggregate Root & Entities)

```go
package user

import (
	"context"
	"time"
)

// User 聚合根
type User struct {
	ID        uint
	Email     Email
	Name      string
	Password  Password
	Avatar    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser 工厂方法
func NewUser(email Email, name string, password Password) *User {
	return &User{
		Email:     email,
		Name:      name,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// 领域行为：更新资料
func (u *User) UpdateProfile(name, avatar string) {
	if name != "" {
		u.Name = name
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	u.UpdatedAt = time.Now()
}
```

#### 仓储与外部服务接口 (Repository & Service Interfaces)

**关键原则**：接口定义在领域层，实现定义在基础设施层。

```go
package user

import "context"

// UserRepository 用户实体仓储 (实现通常是 MySQL)
type Repository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

// VerifyCodeRepository 验证码仓储 (实现通常是 Redis)
type VerifyCodeRepository interface {
    SetCode(ctx context.Context, email string, code string, ttl int) error
    GetCode(ctx context.Context, email string) (string, error)
    DeleteCode(ctx context.Context, email string) error
}

// EmailSender 邮件发送服务 (实现通常是 SMTP / 第三方 API)
// 这也是一个防腐层接口，领域层不关心具体是用 SendGrid 还是 SMTP
type EmailSender interface {
    SendVerifyCode(ctx context.Context, email string, code string) error
}
```

#### 领域服务接口 (Domain Service Interfaces)

```go
package user

import "context"

type DomainService interface {
    // SendRegisterCode 发送注册验证码
    // 实现中会调用 VerifyCodeRepository.SetCode 和 EmailSender.SendVerifyCode
    SendRegisterCode(ctx context.Context, email string) error
    
    // RegisterWithCode 使用验证码注册
    RegisterWithCode(ctx context.Context, email, code, password, name string) (*User, error)
    
    // Login 处理登录校验
    Login(ctx context.Context, email, password string) (*User, error)
}
```

---

### 2.2 比赛领域 (Match Domain)

**Package**: `internal/domain/match`

#### 值对象 (Value Objects)

```go
package match

import "errors"

type RoomStatus int

const (
	RoomStatusPending    RoomStatus = 1 // 未开始
	RoomStatusInProgress RoomStatus = 2 // 进行中
	RoomStatusFinished   RoomStatus = 3 // 已完成
)

type PlayerType int

const (
	PlayerTypeVirtual PlayerType = 1
	PlayerTypeReal    PlayerType = 2
)

// GameConfig 比赛配置值对象
// 仅存储配置信息，不包含计算逻辑
type GameConfig struct {
	GameType    string // 比如 "snooker", "chinese_8"
	MaxPlayers  int
	TargetScore int    // 可选，目标分数（如抢几）
	IsPrivate   bool
	AllowUndo   bool
}

// 默认配置
var DefaultGameConfig = GameConfig{
	GameType:   "generic",
	MaxPlayers: 2,
	IsPrivate:  false,
	AllowUndo:  true,
}
```

#### 聚合根与实体 (Aggregate Root & Entities)

```go
package match

import (
	"context"
	"errors"
	"time"
	"github.com/google/uuid"
)

// Room 聚合根
type Room struct {
	ID        uint
	RoomCode  string
	OwnerID   uint
	Status    RoomStatus
	Config    GameConfig // 使用 GameConfig
	Players   []*Player
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Player 实体 (属于 Room 聚合)
type Player struct {
	ID        uint
	RoomID    uint
	UserID    *uint
	NickName  string
	Type      PlayerType
	IsOnline  bool
	JoinTime  time.Time
}

// 工厂方法
func NewRoom(ownerID uint, config GameConfig) *Room {
	return &Room{
		OwnerID:   ownerID,
		RoomCode:  uuid.New().String()[:6],
		Status:    RoomStatusPending,
		Config:    config,
		Players:   make([]*Player, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// 领域行为
func (r *Room) Start(operatorID uint) error {
	if r.OwnerID != operatorID {
		return errors.New("only owner can start")
	}
	if r.Status != RoomStatusPending {
		return errors.New("room not pending")
	}
	if len(r.Players) < 1 {
		return errors.New("not enough players")
	}
	r.Status = RoomStatusInProgress
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Room) AddPlayer(player *Player) error {
	if len(r.Players) >= r.Config.MaxPlayers {
		return errors.New("room full")
	}
	r.Players = append(r.Players, player)
	r.UpdatedAt = time.Now()
	return nil
}
```

#### 仓储接口 (Repository Interfaces)

```go
package match

import "context"

type Repository interface {
	Save(ctx context.Context, room *Room) error
	FindByID(ctx context.Context, id uint) (*Room, error)
	FindByCode(ctx context.Context, code string) (*Room, error)
}
```

#### 领域服务接口 (Domain Service Interfaces)

```go
package match

import "context"

type DomainService interface {
    // CreateRoom 负责创建房间的业务流程（如生成不重复的 RoomCode）
    CreateRoom(ctx context.Context, ownerID uint, config GameConfig) (*Room, error)
    // JoinRoom 处理加入房间，包括各种校验（房间是否存在、是否满员、是否已在房间内）
    JoinRoom(ctx context.Context, roomCode string, userID uint, nickName string) (*Room, error)
}
```

---

### 2.3 记分领域 (Scoring Domain)

**Package**: `internal/domain/scoring`

**核心调整**：既然计分逻辑在客户端，Scoring 领域退化为纯粹的“账本”记录。它不关心为什么加分（规则），只关心加了多少分（事实）。

#### 值对象 (Value Objects)

```go
package scoring

import "time"

type ScoreChange struct {
	Value int
}

type MatchStats struct {
	TotalScores map[uint]int // PlayerID -> Score
	FoulCounts  map[uint]int
}
```

#### 聚合根 (Aggregate Root)

```go
package scoring

import (
	"context"
	"time"
)

// ScoreEvent 聚合根 (记录每一次分数变动事实)
type ScoreEvent struct {
	ID         uint
	RoomID     uint
	PlayerID   uint
	OperatorID uint
	Change     ScoreChange            // 客户端计算好的分值，如 +1, -4
	Context    map[string]interface{} // 客户端传入的任何元数据 (如：进球颜色、是否犯规、备注)
	CreatedAt  time.Time
	IsUndone   bool
}

func NewScoreEvent(roomID, playerID, operatorID uint, change int, ctx map[string]interface{}) *ScoreEvent {
	return &ScoreEvent{
		RoomID:     roomID,
		PlayerID:   playerID,
		OperatorID: operatorID,
		Change:     ScoreChange{Value: change},
		Context:    ctx,
		CreatedAt:  time.Now(),
	}
}
```

#### 仓储接口 (Repository Interfaces)

```go
package scoring

import "context"

type Repository interface {
	AddEvent(ctx context.Context, event *ScoreEvent) error
	FindByRoomID(ctx context.Context, roomID uint) ([]*ScoreEvent, error)
	RevokeEvent(ctx context.Context, eventID uint) error
	GetLastValidEvent(ctx context.Context, roomID uint) (*ScoreEvent, error)
}
```

#### 领域服务接口 (Domain Service Interfaces)

```go
package scoring

import "context"

type DomainService interface {
    // RecordScore 记录一次得分操作，可能会涉及一些跨聚合的校验（如房间状态）
    RecordScore(ctx context.Context, roomID, playerID, operatorID uint, change int, context map[string]interface{}) error
    // UndoLastOperation 执行撤回逻辑
    UndoLastOperation(ctx context.Context, roomID, operatorID uint) error
}
```

---

## 3. 防腐层设计 (Anti-Corruption Layer / Data Translation)

### 3.1 Interface Layer <-> Application Layer (DTO <-> DTO/Entity)

**组件**: `Assembler` (或 Mapper)
**位置**: `internal/api/assembler`

接口层接收 DTO (Data Transfer Object)，应用层也可能返回 DTO。Assembler 负责将 Request DTO 转为 Command/Entity，或将 Result DTO 转为 Response DTO。

### 3.2 Domain Layer <-> Infrastructure Layer (Entity <-> PO)

**组件**: `Converter` (或 Data Mapper)
**位置**: `internal/infrastructure/persistence/converter` 或直接在 Repository 实现包内。

基础设施层使用 PO (Persistent Object) 也就是 ORM 模型。Converter 负责将纯净的 Domain Entity 转换为带 SQL 标记的 PO，反之亦然。

---

## 4. 事件总线与 Comet 集成 (Event Bus & Comet)

### 4.1 接口定义 (Domain/Shared Layer)

```go
// internal/domain/shared/event_bus.go
type EventBus interface {
    Publish(ctx context.Context, eventType string, payload interface{}) error
    // Subscribe 定义订阅接口，基础设施层去实现具体的监听
    Subscribe(ctx context.Context, channel string) (<-chan []byte, error)
}
```

### 4.2 应用层 Worker (Application Layer)

将 `CometSubscriber` 的逻辑上移至应用层，作为一个 Worker 存在。

```go
// internal/application/worker/comet_subscriber.go

type CometSubscriber struct {
    eventBus     shared.EventBus
    matchApp     *application.MatchAppService
    // ... other app services
}

// Start 启动后台监听任务
func (s *CometSubscriber) Start(ctx context.Context) {
    // 调用基础设施层的订阅能力
    ch, _ := s.eventBus.Subscribe(ctx, "billiard_channel")
    
    for msg := range ch {
        // 在这里进行分发，如果涉及到业务逻辑，调用 AppService
        // 如果只是纯粹的 Socket 推送，也可以调用 Infrastructure 提供的 Pusher
        s.handleMessage(ctx, msg)
    }
}
```

### 4.3 Comet 依赖倒置具体代码 (Dependency Inversion)

通过 **钩子 (Hook)** 机制切断 Comet 对 Application Service 的直接依赖。

**第一步：在基础设施层定义钩子接口**

```go
// internal/infrastructure/comet/session/manager.go

package session

import (
	"context"
	"net/http"
	"github.com/miebyte/goutils/websocketutils"
)

// SessionHook 定义了外部需要实现的回调
// Comet 只知道需要验证 Token，不知道具体怎么验证
type SessionHook interface {
	OnConnect(ctx context.Context, token string) (userID uint, sessionID string, err error)
	OnDisconnect(ctx context.Context, userID uint)
}

type SessionManager struct {
	hook   SessionHook // 依赖接口
	socket websocketutils.ServerAPI
	// ...
}

func NewSessionManager(hook SessionHook) *SessionManager {
	sm := &SessionManager{hook: hook}

	// 在初始化 Socket 时使用 hook
	allowRequestFunc := func(r *http.Request) (any, bool) {
		// 从 Query 或 Header 获取 Token
		token := r.URL.Query().Get("token")
		
		// 调用 Hook 进行校验
		userID, sessionID, err := hook.OnConnect(r.Context(), token)
		if err != nil {
			return nil, false
		}
		
		// 返回认证信息
		return map[string]interface{}{
			"user_id":    userID,
			"session_id": sessionID,
		}, true
	}

	sm.socket = websocketutils.NewServer(
		websocketutils.WithAllowRequestFunc(allowRequestFunc),
		// ...
	)
	return sm
}
```

**第二步：在接口层/应用层实现钩子**

```go
// internal/api/handler/socket_hook.go

package handler

import (
	"context"
	"github.com/superwhys/billiard-helper/internal/application"
)

// AppSessionHook 实现了 Comet 的 SessionHook 接口
type AppSessionHook struct {
	userApp  *application.UserAppService
	matchApp *application.MatchAppService
}

func NewAppSessionHook(userApp *application.UserAppService, matchApp *application.MatchAppService) *AppSessionHook {
	return &AppSessionHook{
		userApp:  userApp,
		matchApp: matchApp,
	}
}

// OnConnect 调用 Application Service 进行 Token 校验
func (h *AppSessionHook) OnConnect(ctx context.Context, token string) (uint, string, error) {
	// 假设 UserAppService 有 VerifyToken 方法
	userDTO, err := h.userApp.VerifyToken(ctx, token)
	if err != nil {
		return 0, "", err
	}
	// 生成一个新的 SessionID (UUID)
	return userDTO.ID, generateUUID(), nil 
}

// OnDisconnect 处理断连逻辑
func (h *AppSessionHook) OnDisconnect(ctx context.Context, userID uint) {
	// 调用 Application Service 处理用户离线（如广播离开房间）
	h.matchApp.HandleUserOffline(ctx, userID)
}
```

**第三步：在 Main 中组装**

```go
// cmd/server/main.go

func main() {
    // 1. 初始化 App Services
    userApp := application.NewUserAppService(...)
    matchApp := application.NewMatchAppService(...)

    // 2. 初始化 Hook
    socketHook := handler.NewAppSessionHook(userApp, matchApp)

    // 3. 初始化 Comet Server (注入 Hook)
    cometServer := comet.NewCometServer(queue, socketHook)
    
    // ...
}
```

### 4.4 JWT 依赖重构 (JWT Dependency)

`internal/pkg/jwt` 不应依赖领域模型，应改为纯粹的 DTO。为了满足在 Token 中存储用户基本信息的需求，我们可以在 `UserClaims` 中定义基础字段。

```go
// internal/pkg/jwt/jwt.go

// UserClaims 定义 JWT 的 Payload 结构
// 关键点：这里只定义基础类型字段，不引用 domain.User 实体，避免循环依赖
type UserClaims struct {
    UserID    uint   `json:"uid"`
    SessionID string `json:"sid"`
    Name      string `json:"name,omitempty"`  // 冗余存储用户信息
    Avatar    string `json:"avatar,omitempty"`
    jwt.RegisteredClaims
}

// GenerateToken 生成 Token (接收 Claims 结构体)
func GenerateToken(claims UserClaims) (string, error) { 
    // ... 使用 jwt 库签名 ...
}

// ParseToken 解析 Token (纯函数，不查库)
func ParseToken(token string) (*UserClaims, error) { ... }
```

**使用方式 (Application Layer)**:
在 `UserAppService` 的登录逻辑中，负责将 `domain.User` 转换为 `jwt.UserClaims`。

```go
// internal/application/user_app.go

func (s *UserAppService) Login(ctx context.Context, email, password string) (string, error) {
    // 1. 调用领域服务进行登录校验
    user, err := s.userDomainService.Login(ctx, email, password)
    if err != nil {
        return "", err
    }

    // 2. 组装 Claims (Entity -> DTO)
    claims := jwt.UserClaims{
        UserID: user.ID,
        Name:   user.Name,
        Avatar: user.Avatar,
        SessionID: generateUUID(), // 生成新的会话ID
    }

    // 3. 生成 Token
    return jwt.GenerateToken(claims)
}
```

**SessionManager 中的 Claims 处理**:
*   **移除** `middlewares.TokenClaimsFromContext` 调用。
*   **Hook 传递**: 利用 `allowRequestFunc` 返回的 map，WebSocket 库通常会将其注入连接上下文。
*   **使用**: `ctx.Conn().GetString("user_id")`。

---

## 5. 推荐目录结构 (Directory Structure)

```text
.
├── cmd
│   └── server
├── internal
│   ├── api                 // [接口层]
│   │   ├── dto             // Data Transfer Objects (Request/Response)
│   │   ├── assembler       // DTO <-> Entity/Command 转换器 (ACL)
│   │   ├── handler         // HTTP Handlers
│   │   │   └── socket_hook.go // [新增] 实现 Comet 的回调接口
│   │   ├── middleware      // Auth Middleware (调用 Application Service)
│   │   └── router
│   │
│   ├── application         // [应用层] (扁平结构，直接存放 App Service)
│   │   ├── user_app.go     // User Application Service
│   │   ├── match_app.go    // Match Application Service (publishes events)
│   │   ├── score_app.go    // Score Application Service
│   │   │
│   │   └── worker          // [新增] 后台任务/消费者
│   │       └── comet_subscriber.go // 监听队列，调度消息推送
│   │
│   ├── domain              // [领域层] (纯净接口定义)
│   │   ├── user            // 定义：UserRepository, VerifyCodeRepository, EmailSender
│   │   ├── match           // 定义：MatchRepository
│   │   ├── scoring         // 定义：ScoringRepository
│   │   └── shared          // 定义：EventBus interface
│   │
│   ├── infrastructure      // [基础设施层] (具体实现)
│   │   ├── persistence     // 持久化实现
│   │   │   ├── db
│   │   │   │   ├── models      // Persistent Objects (GORM Models)
│   │   │   │   └── repo_impl   // Repository 接口实现 (GORM)
│   │   │   ├── cache
│   │   │   │   └── repo_impl   // VerifyCodeRepository 实现 (Redis)
│   │   │   └── converter       // Entity <-> PO 转换器 (ACL)
│   │   │
│   │   ├── email           // EmailSender 实现 (SMTP/ThirdParty)
│   │   │   └── sender_impl.go
│   │   │
│   │   ├── eventbus        // EventBus 实现 (Comet Queue Wrapper)
│   │   │   └── queue_impl.go
│   │   │
│   │   └── comet           // WebSocket Server Implementation
│   │       ├── server.go   // Comet Server Entry
│   │       ├── session     // Session Management (Defines Hooks)
│   │       └── dispatcher  // Message Handlers for Push
│   │
│   └── pkg                 // 通用工具包 (Logger, JWT utils 等)
└── README.md
```
