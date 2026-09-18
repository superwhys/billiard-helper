package match

import (
	"encoding/json"
	"strconv"
)

// ParseGameScores 解析比赛分数快照
func ParseGameScores(scores json.RawMessage) (map[uint]GameScore[map[string]uint], error) {
	if len(scores) == 0 {
		return map[uint]GameScore[map[string]uint]{}, nil
	}

	var scoreMap map[string]GameScore[map[string]uint]
	if err := json.Unmarshal(scores, &scoreMap); err != nil {
		return nil, err
	}

	resp := make(map[uint]GameScore[map[string]uint], len(scoreMap))
	for playerID, score := range scoreMap {
		id, err := strconv.ParseUint(playerID, 10, 64)
		if err != nil {
			continue
		}
		resp[uint(id)] = score
	}

	return resp, nil
}

// FindGameWinner 计算当前局赢家及分数
func FindGameWinner(scoreMap map[uint]GameScore[map[string]uint]) (*uint, int) {
	var winnerID *uint
	var winnerScore int
	hasScore := false
	tie := false

	for playerID, score := range scoreMap {
		if !hasScore || score.Score > winnerScore {
			winnerScore = score.Score
			id := playerID
			winnerID = &id
			hasScore = true
			tie = false
			continue
		}
		if score.Score == winnerScore {
			tie = true
		}
	}

	if !hasScore || tie {
		return nil, 0
	}
	return winnerID, winnerScore
}

// FindMatchWinner 计算比赛赢家及分数
func FindMatchWinner(scoreMap map[uint]int) (*uint, int) {
	var winnerID *uint
	var winnerScore int
	hasScore := false
	tie := false

	for playerID, score := range scoreMap {
		if !hasScore || score > winnerScore {
			winnerScore = score
			id := playerID
			winnerID = &id
			hasScore = true
			tie = false
			continue
		}
		if score == winnerScore {
			tie = true
		}
	}

	if !hasScore || tie {
		return nil, 0
	}
	return winnerID, winnerScore
}

// BuildMatchPlayerScores 计算比赛维度的玩家分数
// 对于九球追分模式和八球模式，只有一局比赛，所以直接计算这一局的分数即可
func BuildMatchPlayerScores(matchType MatchType, games []*MatchGame) (map[uint]int, error) {
	playerScores := map[uint]int{}

	for _, game := range games {
		scoreMap, err := ParseGameScores(game.Scores)
		if err != nil {
			return nil, err
		}

		switch matchType {
		case MatchTypeSnooker:
			for playerID := range scoreMap {
				if _, ok := playerScores[playerID]; !ok {
					playerScores[playerID] = 0
				}
			}
			if game.EndAt != 0 && game.WinnerID != nil {
				playerScores[*game.WinnerID]++
			}
		case MatchType9Ball, MatchType8Ball:
			for playerID, score := range scoreMap {
				playerScores[playerID] = score.Score
			}
		}
	}

	return playerScores, nil
}
