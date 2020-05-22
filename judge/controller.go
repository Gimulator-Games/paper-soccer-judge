package judge

import (
	"encoding/json"
	"fmt"
	"time"

	client "github.com/Gimulator/client-go"
)

const (
	typeVerdict   = "verdict"
	typeAction    = "action"
	typeRegister  = "register"
	typeEndOfGame = "end-of-game"
	namespace     = "paper-soccer"
	worldName     = "world"

	apiTimeWait = 3
)

type controller struct {
	*client.Client
}

func newController(ch chan client.Object) (controller, error) {
	cli, err := client.NewClient(ch)
	if err != nil {
		return controller{}, err
	}

	err = cli.Watch(client.Key{
		Name:      "",
		Namespace: namespace,
		Type:      typeAction,
	})

	if err != nil {
		return controller{}, err
	}

	return controller{
		cli,
	}, nil
}

func (c *controller) setWorld(w World) error {
	val, err := json.Marshal(w)
	if err != nil {
		return err
	}

	fmt.Println("world turn =", w.Turn)

	obj := client.Object{
		Key: client.Key{
			Type:      typeVerdict,
			Namespace: namespace,
			Name:      worldName,
		},
		Value: string(val),
	}

	for {
		err = c.Set(obj)
		if err == nil {
			return nil
		}

		time.Sleep(time.Second * apiTimeWait)
	}
}

func (c *controller) setEndOfGame(winner string) error {
	fmt.Println("end of game =", winner)
	obj := client.Object{
		Key: client.Key{
			Type:      typeEndOfGame,
			Namespace: namespace,
			Name:      "",
		},
		Value: winner,
	}

	for {
		err := c.Set(obj)
		if err == nil {
			return nil
		}

		time.Sleep(time.Second * apiTimeWait)
	}
}

func (c *controller) receiptPlayers() (client.Object, client.Object) {
	for {
		objs, err := c.Find(client.Key{
			Name:      "",
			Namespace: namespace,
			Type:      typeRegister,
		})
		if err != nil {
			continue
		}

		if len(objs) == 2 {
			return objs[0], objs[1]
		}

		time.Sleep(time.Second * 3)
	}
}
