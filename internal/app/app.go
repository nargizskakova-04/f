package app

import (
	"fmt"

	"frappuccino/internal/config"
	"frappuccino/internal/server"
)

const cfgPath = "./config/config.json"

func Start() {
	config, err := config.GetConfig(cfgPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	app := server.NewApp(config)
	if err := app.Initialize(); err != nil {
		fmt.Println("Error during start: ", err)
	}
	app.Run()
}
