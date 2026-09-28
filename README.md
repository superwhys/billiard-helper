# billiard-helper

## 业务逻辑

### 创建对局

1. 创建 Match
2. 遍历创建 Player
3. 创建对局的第一轮 MatchGame(round=1)

### 记分操作同步

1. 获取比赛 Match
2. 获取比赛轮 MatchGame
3. 创建 Event 记录
4. 计算当前轮的分数
    - 获取当前的分数快照
    - 遍历当前事件的 ScoreActions，根据 ScoreActions 中的 PlayerIds 和 Score 计算新的分数快照
    - 这里需要使用策略模式，不同的玩法使用不同的计算策略，最终返回一个分数快照即可(不同玩法的分数快照数据结构可能不同)
5. 更新 MatchGame 的分数快照和最近处理的一次事件 ID

### 撤回记分操作

1. 获取比赛 Match
2. 获取比赛轮 MatchGame
3. 获取当前 MatchGame 对应的最新的一个 Event
4. 删除该事件
5. 根据当前 MatchGame 分数快照，撤回当前事件的应用 (这里也需要一个策略模式，计算出撤回后的分数快照)
6. 获取当前比赛轮的最新的 event
7. 更新 MatchGame 的分数快照和最新的事件 ID
## 斯诺克（Web）

- 双人记分器，支持 1–35 的奇数局赛制（如 3 局 2 胜），红球数量在 `config.data.red_count` 中配置，范围 1–15，默认 15；有效进球按红彩交替和清彩顺序校验。
- 红、黄、绿、棕、蓝、粉、黑球分别记 1–7 分；犯规给对手加 4–7 分，犯规者不扣分。
- `/score/sync` 沿用现有协议：`context.stat_key` 为 `red/yellow/green/brown/blue/pink/black/foul`，`context.scorer_player_id` 为进球或犯规球员；`score_actions` 仅允许一个对应分值和一个收分球员。后端校验球员归属、分值与收分方向。
- `/score/sync` 的斯诺克快照在数字球员 ID 外包含 `_snooker` 元数据：`red_count`、`reds_remaining`、`next_ball`、`active_player_id`、`break_score`、`remaining_points`、`can_undo`。红球后任选彩球，最后红球后的彩球机会结束后按黄到黑清彩；清台平分重置黑球。
- 换人使用 `context.stat_key=turn_end`、当前 `scorer_player_id`、对手 `next_player_id`，`score_actions` 为空。犯规可传 `reds_removed`（实际离台红球数量）及 `next_player_id`（默认对手，也可要求犯规者继续）。换人和犯规结束当前单杆，罚分不计入单杆。
- `/score/undo` 撤销当前局最新一笔。服务器在事件顶层写入 `before_scores`，完整恢复球序、击球球员、红球数、分数和单杆；客户端无法指定恢复快照。
- 理论单杆满分为 `红球数 × 8 + 27`。台面剩余分数在等红球时为 `剩余红球 × 8 + 27`，等彩球时加 7，清彩时计算剩余彩球分值；不含未来罚分或自由球。
- `/match/round/next` 结算斯诺克本局：提交 `match_id`、`round`，认输时增加 `conceding_player_id`。平分不能直接结算；认输保留实际比分。达到多数胜局自动完成整场，否则新建下一局。
- 结算后不能撤销上一局。局胜者保存在已有 `match_games.winner_id`，整场成绩按已完成的局胜者汇总，无需新增数据库字段。
- 斯诺克计分、撤销与结算在事务中锁定比赛记录，防止并发覆盖；过期局号的请求被拒绝。`/match/end` 不用于斯诺克结算。
- 计分依据：[WPBSA 规则](https://www.wpbsa.com/rules/)。支持基础球序与台面计数，暂不自动处理自由球、多红球同杆等特殊裁判场景。

验证：`go test ./...`、`go vet ./...`。真实 MySQL / Redis 集成测试使用独立测试库和实例（会建表及写入测试数据），显式设置以下环境变量后运行：

```sh
BILLIARD_TEST_MYSQL_DSN='<独立测试库 DSN，含 parseTime=true>' \
BILLIARD_TEST_REDIS_ADDR='127.0.0.1:<独立测试 Redis 端口>' \
go test ./internal/integration -run TestSnookerHTTPPersistence -v
```

升级前的旧局缺少台面状态时保留原手动记分；下一局使用新规则。新局的状态直接存入现有 JSON 分数快照，无需新增数据库字段。

## 用户黑名单

服务启动时会创建 `blocked_identities` 表。部署新版服务后，在服务器上使用与服务相同的配置文件执行管理命令：

```sh
go run ./cmd/blacklist -f /path/to/config.json --action block --kind email --value 'someone@example.com'
go run ./cmd/blacklist -f /path/to/config.json --action block --kind open_id --value 'wechat-open-id'
go run ./cmd/blacklist -f /path/to/config.json --action block --kind user_id --value '123'
go run ./cmd/blacklist -f /path/to/config.json --action unblock --kind email --value 'someone@example.com'
```

`block` 和 `unblock` 可重复执行。拉黑后，登录、注册、验证码、已有访问令牌和刷新令牌均被拒绝；命令还会通知服务断开该用户现有的 WebSocket 连接。若命令提示黑名单已保存但断线通知失败，可重复执行相同的 `block` 命令。

## Web 超管管理

在服务配置文件现有的 `config` 段中加入超管用户 ID，保留其他配置，重启服务后生效：

```json
{
  "config": {
    "super_admin_user_id": 123
  }
}
```

不配置或设为 `0` 时关闭超管权限。超管身份只由此配置决定，不能通过前端或数据库用户字段设置。

Web 启动时请求 `GET /api/app/state` 获取状态配置，其中 `is_super_admin` 表示当前用户是否为超管；匿名访问返回 `false`，携带令牌时验证会话及黑名单。超管在“我的 → 常用设置”中看到“超管管理”入口。

- `GET /api/admin/users?page=1&page_size=20`：分页展示用户及黑名单状态，不返回密码。每页最多 100 条。
- `POST /api/admin/users/block`：提交 `{"user_id":123}` 拉黑目标账号，立即拦截其旧令牌并通知所有服务实例断开 WebSocket。禁止在页面中拉黑超管自身；已有邮箱和 OpenID 黑名单继续生效。页面暂不提供解封，运维可使用上面的命令。
- `GET /api/admin/matches/stats`：按北京时间当天 `[00:00, 次日00:00)` 统计创建的比赛总数及玩法分布，包含已软删除的比赛。

所有 `/api/admin` 接口都校验有效登录态和配置中的超管身份；隐藏前端入口不替代服务端权限检查。

未设置上述变量时只跳过外部服务集成测试。该测试经过真实 HTTP Handler、应用服务、计分策略及数据库事务，覆盖并发记分、撤销、非法请求回滚、多局结算和成绩重载。
