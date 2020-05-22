package judge

import (
	"encoding/json"
	"math"
	"os"
	"sync"
	"time"

	client "github.com/Gimulator/client-go"
	uuid "github.com/satori/go.uuid"
)

type moveResult string

const (
	validMove   moveResult = "valid-move"
	invalidMove moveResult = "invalid-move"
	prizeMove   moveResult = "prize-move"
	winningMove moveResult = "winning-move"
	losingMove  moveResult = "losing-move"

	width  = 11
	height = 15
)

type Judge struct {
	*sync.Mutex
	controller

	world       World
	playground  [][]int
	token       string
	ownerToName map[string]string
	ch          chan client.Object
}

func NewJudge() (Judge, error) {
	ch := make(chan client.Object, 16)
	c, err := newController(ch)
	if err != nil {
		return Judge{}, err
	}

	judge := Judge{
		Mutex:       &sync.Mutex{},
		controller:  c,
		ownerToName: make(map[string]string),
		ch:          ch,
	}

	if err = judge.load(); err != nil {
		return Judge{}, err
	}
	return judge, nil
}

func (j *Judge) load() error {
	obj1, obj2 := j.receiptPlayers()

	o1 := obj1.Owner
	p1 := obj1.Key.Name
	j.ownerToName[o1] = p1

	o2 := obj2.Owner
	p2 := obj2.Key.Name
	j.ownerToName[o2] = p2

	w, err := NewWorld(p1, p2, width, height)
	if err != nil {
		return err
	}
	j.world = w

	j.playground = genPlayground(w)

	if err := j.setWorld(j.world); err != nil {
		return err
	}

	return nil
}

func (j *Judge) Listen() {
	for {
		obj := <-j.ch
		if obj.Key.Type != typeAction || obj.Key.Name != j.ownerToName[obj.Owner] {
			continue
		}

		var move Move
		err := json.Unmarshal([]byte(obj.Value.(string)), &move)
		if err != nil {
			continue
		}

		j.judge(move, obj.Key.Name)
	}
}

func (j *Judge) judge(move Move, name string) {
	j.Lock()
	defer j.Unlock()

	if name == j.world.Player1.Name {
		move.Player = j.world.Player1
	} else if name == j.world.Player2.Name {
		move.Player = j.world.Player2
	} else {
		return
	}

	if j.world.Turn != move.Player.Name {
		return
	}

	res := j.judgeMove(move)
	j.update(move, res)
}

func (j *Judge) timer(token string) {
	time.Sleep(3 * time.Second)

	j.Lock()
	defer j.Unlock()

	if token != j.token {
		return
	}

	j.update(Move{}, invalidMove)
}

func (j *Judge) judgeMove(move Move) moveResult {
	validMoves := j.validMoves()
	if j.isInvalidMove(move, validMoves) {
		return invalidMove
	}

	if j.isWinningMove(move) {
		return winningMove
	}

	if j.isLosingMove(move) {
		return losingMove
	}

	if j.isBlockingMove(move) {
		return losingMove
	}

	if j.isPrizeMove(move) {
		return prizeMove
	}

	return validMove
}

func (j *Judge) isInvalidMove(move Move, valids []Move) bool {
	for _, m := range valids {
		if m.Equal(move) {
			return false
		}
	}
	return true
}

func (j *Judge) isBlockingMove(move Move) bool {
	if j.playground[move.To.X][move.To.Y] >= 7 {
		return true
	}
	return false
}

func (j *Judge) isPrizeMove(move Move) bool {
	if j.playground[move.To.X][move.To.Y] > 0 {
		return true
	}
	return false
}

func (j *Judge) isWinningMove(move Move) bool {
	side := move.Player.Side
	if side == topSide {
		if move.To.Y == 0 && math.Abs(float64(move.To.X-j.world.Width/2)) < 2 {
			return true
		}
		return false
	}
	if move.To.Y == j.world.Height-1 && math.Abs(float64(move.To.X-j.world.Width/2)) < 2 {
		return true
	}
	return false
}

func (j *Judge) isLosingMove(move Move) bool {
	side := move.Player.Side
	if side == downSide {
		if move.To.Y == 0 && math.Abs(float64(move.To.X-j.world.Width/2)) < 2 {
			return true
		}
		return false
	}
	if move.To.Y == j.world.Height-1 && math.Abs(float64(move.To.X-j.world.Width/2)) < 2 {
		return true
	}
	return false
}

func (j *Judge) update(move Move, result moveResult) {
	j.updateToken()
	j.updateTurn(result)

	if result != invalidMove {
		j.world.Moves = append(j.world.Moves, move)
		j.playground[move.To.X][move.To.Y]++
		j.world.BallPos = move.To
	}

	j.setWorld(j.world)

	if result == losingMove {
		for owner, name := range j.ownerToName {
			if name != j.world.Turn {
				j.setEndOfGame(owner)
				os.Exit(0)
			}
		}
	}

	if result == winningMove {
		for owner, name := range j.ownerToName {
			if name == j.world.Turn {
				j.setEndOfGame(owner)
				os.Exit(0)
			}
		}
	}
	go j.timer(j.token)
}

func (j *Judge) updateToken() {
	j.token = uuid.NewV4().String()
}

func (j *Judge) updateTurn(res moveResult) {
	switch res {
	case invalidMove:
		j.changeTurn()
	case validMove:
		j.changeTurn()
	case winningMove:
		j.world.Turn = ""
	case losingMove:
		j.world.Turn = ""
	case prizeMove:
		// Nothing to do
	}
}

func (j *Judge) changeTurn() {
	if j.world.Turn == j.world.Player1.Name {
		j.world.Turn = j.world.Player2.Name
	} else {
		j.world.Turn = j.world.Player1.Name
	}
}

var (
	dirX = []int{1, 1, 0, -1, -1, -1, 0, 1}
	dirY = []int{0, 1, 1, 1, 0, -1, -1, -1}
)

func (r *Judge) validMoves() []Move {
	var validMoves []Move

	for ind := 0; ind < 8; ind++ {
		x := r.world.BallPos.X + dirX[ind]
		y := r.world.BallPos.Y + dirY[ind]
		if x < 0 || x >= r.world.Width || y < 0 || y >= r.world.Height {
			continue
		}

		validMove := Move{
			From: r.world.BallPos,
			To: Position{
				X: x,
				Y: y,
			},
		}

		isValid := true
		for _, m := range r.world.Moves {
			if validMove.Equal(m) {
				isValid = false
			}
		}

		for _, m := range r.world.FilledMoves {
			if validMove.Equal(m) {
				isValid = false
			}
		}

		if isValid {
			validMoves = append(validMoves, validMove)
		}
	}
	return validMoves
}

func genPlayground(w World) [][]int {
	var playground = make([][]int, w.Width)
	for i := 0; i < w.Width; i++ {
		playground[i] = make([]int, w.Height)
	}

	for _, move := range w.FilledMoves {
		a := move.From
		b := move.To
		playground[a.X][a.Y]++
		playground[b.X][b.Y]++
	}

	return playground
}
