package main

import (
	"file-server/config"
	fileserver "file-server/file-server"

	"github.com/labstack/echo/v4"
)

func main() {

	envs, err := config.ParseEnvFiles(true, ".env")

	if err != nil {
		panic(err.Error())
	}

	for key, val := range envs {
		println(key + "=" + val)
	}

	if 1 == 1 {
		return
	}

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
