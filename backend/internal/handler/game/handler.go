package game

import "github.com/yorukot/blind-party/internal/schema"

type GameHandler struct {
	GameData map[string]*schema.Game
}