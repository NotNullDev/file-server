package fileserver

import (
	"file-server/config"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

const (
	// 	API_KEY       = "xl61Lfdm6n014u1gn2p8CLBc0P9WFz02f5BcBolvl0k="
	MAX_FILE_SIZE = 3 * 1024 * 1024

// PORT          = 4500
)

type FileServer struct {
	Echo   *echo.Echo
	Config *config.AppConfig
}

func (f *FileServer) InitRoutes() {
	f.Echo.POST("/", receiveFiles)
	f.Echo.GET("/:fileName", getFile)
}

func (f *FileServer) Start() error {
	return f.Echo.Start(fmt.Sprintf(":%d", config.GlobalAppConfig.Port))
}

func receiveFiles(c echo.Context) error {
	if !authorizeUserWithNextAuthServer(&c) {
		return c.JSON(401, ErrorResponse{
			Error: "Could not find session associated with current request.",
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

func authorizeUserWithNextAuthServer(c *echo.Context) bool {
	ct := *c

	client := &http.Client{}
	req, _ := http.NewRequest("GET", config.GlobalAppConfig.AuthServerUrl, nil)
	req.Header = ct.Request().Header

	for _, cookie := range ct.Request().Cookies() {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != 200 {
		return false
	}

	bodyContent, _ := io.ReadAll(resp.Body)

	if string(bodyContent) == "{}" {
		return false
	}

	log.Println(string(bodyContent))

	return true
}
