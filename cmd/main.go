package main

import (
	fileserver "file-server/file-server"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	fileServer := fileserver.FileServer{Echo: e}
	fileServer.InitRoutes()

	err := fileServer.Start()

	if err != nil {
		panic(err)
	}
}
