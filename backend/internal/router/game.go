package router

import (
	"github.com/go-chi/chi/v5"
	"golang.org/x/net/websocket"

	"github.com/yorukot/blind-party/internal/handler/game"
	"github.com/yorukot/blind-party/internal/schema"
)

// GameRouter sets up the game routes
func GameRouter(r chi.Router) {

	gameHandler := &game.GameHandler{
		GameData: make(map[string]*schema.Game),
	}

	r.Route("/game", func(r chi.Router) {
		r.Post("/", gameHandler.NewGame)
		r.Get("/{gameID}/state", gameHandler.GetGameState)
		r.Route("/{gameID}", func(r chi.Router) {
			r.Handle("/ws", websocket.Handler(gameHandler.ConnectWebSocket))
		})
	})
}
