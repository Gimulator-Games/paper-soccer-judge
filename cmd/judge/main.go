package main

import (
	"math/rand"
	"time"

	"github.com/Gimulator-Games/paper-soccer-judge/judge"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	j, err := judge.NewJudge()
	if err != nil {
		panic(err)
	}

	j.Listen()
}
