package main

import (
	"miopkg/application"

	"miopkg/conf"
	"miopkg/log"
)

//  go run main.go --config=config.toml
func main() {
	app := NewEngine()
	if err := app.Run(); err != nil {
		panic(err)
	}
}

type Engine struct {
	application.Application
}

func NewEngine() *Engine {
	eng := &Engine{}
	if err := eng.Startup(
		eng.printConfig,
	); err != nil {
		log.Panic("startup", log.Any("err", err))
	}
	return eng
}

func (s *Engine) printConfig() error {
	log.DefaultLogger = log.StdConfig("default").Build()
	peopleName := conf.GetString("people.name")
	log.Info("people info", log.String("name", peopleName), log.String("type", "onelineByFile"))
	return nil
}
