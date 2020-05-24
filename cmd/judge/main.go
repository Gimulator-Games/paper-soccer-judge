package main

import (
	"fmt"
	"math/rand"
	"path"
	"runtime"
	"time"

	"github.com/Gimulator-Games/paper-soccer-judge/judge"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)

	formatter := &logrus.TextFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		FullTimestamp:    true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		ForceQuote:       false,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf(" %s:%d\t", path.Base(f.File), f.Line)
		},
	}
	logrus.SetFormatter(formatter)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	j, err := judge.NewJudge()
	if err != nil {
		panic(err)
	}

	j.Listen()
}
