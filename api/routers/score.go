package routers

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
)

func ScoreGroupRouter() ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/score"),
		ginutils.WithHandler(http.MethodPost, "/sync", nil),
		ginutils.WithHandler(http.MethodPost, "/undo", nil),
		ginutils.WithHandler(http.MethodGet, "/list", nil),
	)
}
