package main

import (
	"time"

	"github.com/holoplot/kubelish/cmd/kubelish/cmd"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        colorable.NewColorableStdout(),
		TimeFormat: time.RFC3339,
	}

	log.Logger = log.Output(consoleWriter)

	cmd.Execute()
}
