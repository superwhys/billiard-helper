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