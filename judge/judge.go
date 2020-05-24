package judge

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	client "github.com/Gimulator/client-go"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
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
	log         *logrus.Entry
}

func NewJudge() (*Judge, error) {
	ch := make(chan client.Object, 16)
	c, err := newController(ch)
	if err != nil {
		return nil, err
	}

	judge := &Judge{
		Mutex:       &sync.Mutex{},
		controller:  c,
		ownerToName: make(map[string]string),
		ch:          ch,
		log:         logrus.WithField("entity", "judge"),
	}

	if err = judge.load(); err != nil {
		return nil, err
	}
	return judge, nil
}

func (j *Judge) load() error {
	j.log.Info("starting to load game")
	obj1, obj2 := j.receiptPlayers()

	o1 := obj1.Owner
	p1 := obj1.Key.Name
	j.ownerToName[o1] = p1

	o2 := obj2.Owner
	p2 := obj2.Key.Name
	j.ownerToName[o2] = p2
	j.log.WithFields(logrus.Fields{
		"player1": p1,
		"player2": p2,
		"owner1":  o1,
		"owner2":  o2,
	}).Info("receipted players")

	w, err := NewWorld(p1, p2, width, height)
	if err != nil {
		j.log.WithError(err).Error("could not initiate world")
		return err
	}
	j.world = w

	j.playground = genPlayground(w)

	if err := j.setWorld(j.world); err != nil {
		j.log.WithError(err).Fatal("could not set world")
		return err
	}

	return nil
}

func (j *Judge) Listen() {
	for {
		obj := <-j.ch
		log := j.log.WithField("object", obj)

		log.Debug("receiving new object from gimulator")
		if obj.Key.Type != typeAction || obj.Key.Name != j.ownerToName[obj.Owner] {
			log.Error("invalid action type or name")
			continue
		}

		var move Move
		err := json.Unmarshal([]byte(obj.Value.(string)), &move)
		if err != nil {
			log.Error("could not unmarshal move in value of object")
			continue
		}

		j.judge(move, obj.Key.Name)
	}
}

func (j *Judge) judge(move Move, name string) {
	log := j.log.WithField("move", move)

	j.Lock()
	defer j.Unlock()

	if name == j.world.Player1.Name {
		move.Player = j.world.Player1
	} else if name == j.world.Player2.Name {
		move.Player = j.world.Player2
	} else {
		log.Fatal("invalid name for move")
		return
	}

	if j.world.Turn != move.Player.Name {
		log.Debug("invalid turn for move")
		return
	}

	res := j.judgeMove(move)
	log.WithField("result", res).Debug("result of move")

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
		if move.To.Y == 0 && move.To.X-j.world.Width/2 <= 1 && move.To.X-j.world.Width/2 >= -1 {
			return true
		}
		return false
	}
	if move.To.Y == j.world.Height-1 && move.To.X-j.world.Width/2 <= 1 && move.To.X-j.world.Width/2 >= -1 {
		return true
	}
	return false
}

func (j *Judge) isLosingMove(move Move) bool {
	side := move.Player.Side
	if side == downSide {
		if move.To.Y == 0 && move.To.X-j.world.Width/2 <= 1 && move.To.X-j.world.Width/2 >= -1 {
			return true
		}
		return false
	}
	if move.To.Y == j.world.Height-1 && move.To.X-j.world.Width/2 <= 1 && move.To.X-j.world.Width/2 >= -1 {
		return true
	}
	return false
}

func (j *Judge) update(move Move, result moveResult) {
	j.updateToken()

	turn := j.world.Turn
	j.updateTurn(result)

	if result != invalidMove {
		j.world.Moves = append(j.world.Moves, move)
		j.playground[move.To.X][move.To.Y]++
		j.world.BallPos = move.To
	}

	j.setWorld(j.world)
	j.log.Debug("set world")

	if result == losingMove {
		for owner, name := range j.ownerToName {
			if name != turn {
				j.setEndOfGame(owner)
				j.log.WithField("winner", owner).Debug("set end of game")
				os.Exit(0)
			}
		}
	}

	if result == winningMove {
		for owner, name := range j.ownerToName {
			if name == turn {
				j.setEndOfGame(owner)
				j.log.WithField("winner", owner).Debug("set end of game")
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
