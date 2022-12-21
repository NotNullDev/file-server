package fileserver

import (
	"fmt"
	"io"
	"os"

	"github.com/labstack/echo/v4"
)

const (
	API_KEY       = ""
	MAX_FILE_SIZE = 3 * 1024 * 1024
	PORT          = 4500
)

type FileServer struct {
	Echo *echo.Echo
}

func (f *FileServer) InitRoutes() {
	f.Echo.POST("/", receiveFiles)
	f.Echo.GET("/:fileName", getFile)
}

func (f *FileServer) Start() error {
	return f.Echo.Start(fmt.Sprintf(":%d", PORT))
}

func receiveFiles(c echo.Context) error {
	apiKey := c.Request().Header.Get("API_KEY")

	if apiKey != API_KEY {
		return c.JSON(401, ErrorResponse{
			Error: "Invalid API key.",
		})
	}

	form, err := c.MultipartForm()

	if err != nil {
		return c.JSON(400, ErrorResponse{
			Error: "You must provide files.",
		})
	}

	files := form.File["files"]

	for _, file := range files {
		// TODO: add saved files rollback in array.
		if file.Size > MAX_FILE_SIZE {
			return c.JSON(400, ErrorResponse{
				Error: "Max file size exceeded.",
			})
		}

		dataStream, err := file.Open()

		if err != nil {
			return c.JSON(400, ErrorResponse{
				Error: "Unknown error.",
			})
		}
		defer dataStream.Close()

		newFile, err := os.Create(file.Filename)

		if err != nil {
			return c.JSON(400, ErrorResponse{
				Error: "Unknown error.",
			})
		}

		_, err = io.Copy(newFile, dataStream)

		if err != nil {
			return c.JSON(400, ErrorResponse{
				Error: "Unknown error.",
			})
		}
	}

	return c.JSON(200, len(files))
}

func getFile(c echo.Context) error {
	fileName := c.Param("fileName")
	if fileName == "" {
		c.JSON(400, ErrorResponse{
			Error: "missing filename",
		})
	}
	return c.File(fileName)
}

type ErrorResponse struct {
	Error string
}
