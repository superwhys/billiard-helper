package match

type CreationStats struct {
	Total  int64
	ByType map[MatchType]int64
}
