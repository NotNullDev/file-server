package main

import (
	"file-server/config"
	fileserver "file-server/file-server"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	config, err := config.NewAppConfigFromEnv()

	if err != nil {
		panic(err.Error())
	}

	fileServer := fileserver.FileServer{Echo: e, Config: &config}
	fileServer.InitRoutes()

	err = fileServer.Start()

	if err != nil {
		panic(err)
	}
}
