package judge

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	topSide  = "top"
	downSide = "down"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p *Position) Equal(pos Position) bool {
	if p.X == pos.X && p.Y == pos.Y {
		return true
	}
	return false
}

type Move struct {
	Player Player
	From   Position `json:"from"`
	To     Position `json:"to"`
}

func (m Move) String() string {
	return fmt.Sprintf("{Player: %s, From: %v, To: %v }", m.Player.Name, m.From, m.To)
}

func (m *Move) Equal(move Move) bool {
	if m.From.Equal(move.From) && m.To.Equal(move.To) {
		return true
	}
	if m.From.Equal(move.To) && m.To.Equal(move.From) {
		return true
	}
	return false
}

type Player struct {
	Name string `json:"name"`
	Side string `json:"side"`
}

func NewPlayer(name string, side string) Player {
	return Player{
		Name: name,
		Side: side,
	}
}

type World struct {
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	Moves       []Move   `json:"moves"`
	FilledMoves []Move   `json:"filled-moves"`
	Turn        string   `json:"turn"`
	BallPos     Position `json:"ball-pos"`
	Player1     Player   `json:"player1"`
	Player2     Player   `json:"player2"`
}

func NewWorld(playerName1, playerName2 string, width, height int) (World, error) {
	if width%2 == 0 || height%2 == 0 {
		return World{}, fmt.Errorf("The height and width of the world must be odd")
	}

	rand.Seed(time.Now().UnixNano())
	rnd := rand.Intn(2)

	var player1 Player
	var player2 Player
	if rnd == 1 {
		player1 = NewPlayer(playerName1, topSide)
		player2 = NewPlayer(playerName2, downSide)
	} else {
		player1 = NewPlayer(playerName1, downSide)
		player2 = NewPlayer(playerName2, topSide)
	}

	world := World{
		Width:       width,
		Height:      height,
		Moves:       make([]Move, 0),
		FilledMoves: GenerateFilledMoves(width, height),
		Turn:        player1.Name,
		BallPos:     Position{X: width / 2, Y: height / 2},
		Player1:     player1,
		Player2:     player2,
	}

	return world, nil
}

func GenerateFilledMoves(width, height int) []Move {
	moves := make([]Move, 0)

	for x := 0; x < width-1; x++ {
		if x-width/2 <= 0 && x-width/2 >= -1 {
			continue
		}
		moves = AddSquareWithDownLeftPos(moves, Position{X: x, Y: height - 2})
		moves = AddSquareWithDownLeftPos(moves, Position{X: x, Y: 0})
	}

	for y := 0; y < height-1; y++ {
		moves = AddSquareWithDownLeftPos(moves, Position{X: 0, Y: y})
		moves = AddSquareWithDownLeftPos(moves, Position{X: width - 2, Y: y})
	}

	moves = append(moves,
		Move{From: Position{X: width / 2, Y: 0}, To: Position{X: width/2 - 1, Y: 0}},
		Move{From: Position{X: width / 2, Y: 0}, To: Position{X: width/2 + 1, Y: 0}},
		Move{From: Position{X: width / 2, Y: height - 1}, To: Position{X: width/2 - 1, Y: height - 1}},
		Move{From: Position{X: width / 2, Y: height - 1}, To: Position{X: width/2 + 1, Y: height - 1}},
	)

	return moves
}

func AddSquareWithDownLeftPos(moves []Move, pos Position) []Move {
	//fmt.Printf("{ %d , %d } ----> { %d , %d }\n", pos.X, pos.Y, pos.X+1, pos.Y)
	//fmt.Printf("{ %d , %d } ----> { %d , %d }\n", pos.X, pos.Y, pos.X, pos.Y+1)
	//fmt.Printf("{ %d , %d } ----> { %d , %d }\n", pos.X, pos.Y, pos.X+1, pos.Y+1)
	//fmt.Printf("{ %d , %d } ----> { %d , %d }\n", pos.X, pos.Y+1, pos.X+1, pos.Y+1)
	//fmt.Printf("{ %d , %d } ----> { %d , %d }\n", pos.X, pos.Y+1, pos.X+1, pos.Y)
	//fmt.Printf("{ %d , %d } ----> { %d , %d }\n", pos.X+1, pos.Y+1, pos.X+1, pos.Y)
	moves = append(moves,
		Move{From: Position{X: pos.X, Y: pos.Y}, To: Position{X: pos.X + 1, Y: pos.Y}},
		Move{From: Position{X: pos.X, Y: pos.Y}, To: Position{X: pos.X, Y: pos.Y + 1}},
		Move{From: Position{X: pos.X, Y: pos.Y}, To: Position{X: pos.X + 1, Y: pos.Y + 1}},
		Move{From: Position{X: pos.X, Y: pos.Y + 1}, To: Position{X: pos.X + 1, Y: pos.Y + 1}},
		Move{From: Position{X: pos.X, Y: pos.Y + 1}, To: Position{X: pos.X + 1, Y: pos.Y}},
		Move{From: Position{X: pos.X + 1, Y: pos.Y + 1}, To: Position{X: pos.X + 1, Y: pos.Y}},
	)
	return moves
}
