package game

import (
	"log"
	"sync"
)

type Game struct {
	Board         []string
	CurrentPlayer string
	Mutex         sync.Mutex
}

func NewGame() *Game {
	game := &Game{
		Board:         make([]string, 64),
		CurrentPlayer: "black",
	}
	game.InitializeBoard()
	return game
}

func (g *Game) InitializeBoard() {
	for i := range g.Board {
		g.Board[i] = ""
	}
	g.Board[27] = "white"
	g.Board[28] = "black"
	g.Board[35] = "black"
	g.Board[36] = "white"
	log.Println("Tabuleiro inicializado.")
}

func (g *Game) HandleMove(index int) bool {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if !g.IsValidMove(index) {
		return false
	}
	g.MakeMove(index)
	g.CurrentPlayer = g.GetOpponent(g.CurrentPlayer)
	return true
}

func (g *Game) IsValidMove(index int) bool {
	if g.Board[index] != "" {
		return false
	}
	directions := []struct {
		x, y int
	}{
		{0, 1}, {1, 1}, {1, 0}, {1, -1},
		{0, -1}, {-1, -1}, {-1, 0}, {-1, 1},
	}
	for _, direction := range directions {
		if g.CapturesInDirection(index, direction) {
			return true
		}
	}
	return false
}

func (g *Game) CapturesInDirection(index int, direction struct{ x, y int }) bool {
	opponent := g.GetOpponent(g.CurrentPlayer)
	hasOpponentBetween := false
	size := 8
	x := index % size
	y := index / size
	x += direction.x
	y += direction.y
	i := y*size + x
	for x >= 0 && x < size && y >= 0 && y < size {
		if g.Board[i] == opponent {
			hasOpponentBetween = true
		} else if g.Board[i] == g.CurrentPlayer {
			return hasOpponentBetween
		} else {
			return false
		}
		x += direction.x
		y += direction.y
		i = y*size + x
	}
	return false
}

func (g *Game) MakeMove(index int) {
	g.FlipCells(index)
	g.Board[index] = g.CurrentPlayer
}

func (g *Game) FlipCells(index int) {
	directions := []struct {
		x, y int
	}{
		{0, 1}, {1, 1}, {1, 0}, {1, -1},
		{0, -1}, {-1, -1}, {-1, 0}, {-1, 1},
	}
	for _, direction := range directions {
		if g.CapturesInDirection(index, direction) {
			g.FlipInDirection(index, direction)
		}
	}
}

func (g *Game) FlipInDirection(index int, direction struct{ x, y int }) {
	opponent := g.GetOpponent(g.CurrentPlayer)
	size := 8
	x := index % size
	y := index / size
	x += direction.x
	y += direction.y
	i := y*size + x
	for x >= 0 && x < size && y >= 0 && y < size {
		if g.Board[i] == opponent {
			g.Board[i] = g.CurrentPlayer
		} else {
			break
		}
		x += direction.x
		y += direction.y
		i = y*size + x
	}
}

func (g *Game) GetOpponent(player string) string {
	if player == "black" {
		return "white"
	}
	return "black"
}

func (g *Game) DetermineWinner() string {
	blackCount, whiteCount := 0, 0
	for _, piece := range g.Board {
		if piece == "black" {
			blackCount++
		} else if piece == "white" {
			whiteCount++
		}
	}
	if blackCount > whiteCount {
		return "Preto vence!"
	} else if whiteCount > blackCount {
		return "Branco vence!"
	}
	return "Empate!"
}
