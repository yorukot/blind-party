package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/yorukot/blind-party/internal/handler/game"
)

// GameRouter sets up the game routes
func GameRouter(r chi.Router) {

	r.Route("/game", func(r chi.Router) {
		r.Route("/{gameID}", func(r chi.Router) {
			r.Post("/join", game.NewGame)
		})
	})
}
