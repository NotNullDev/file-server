package fileserver

import (
	_ "embed"
	"file-server/config"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	MAX_FILE_SIZE = 3 * 1024 * 1024
	FILES_FOLDER  = "files"
)

//go:embed private-public.pem
var cert []byte

type FileServer struct {
	Echo   *echo.Echo
	Config *config.AppConfig
}

func (f *FileServer) InitRoutes() {
	f.Echo.POST("/", receiveFiles)
	f.Echo.GET("/:fileName", getFile)
}

func (f *FileServer) Start() error {
	os.Mkdir(FILES_FOLDER, 0777)

	f.Echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
	}))

	return f.Echo.Start(fmt.Sprintf(":%d", config.GlobalAppConfig.Port))
}

func receiveFiles(c echo.Context) error {
	println(fmt.Sprintf("Cert: [%s]", cert))
	// if !authorizeUserWithNextAuthServer(&c) {
	// 	return c.JSON(401, ErrorResponse{
	// 		Error: "Could not find session associated with current request.",
	// 	})
	// }

	form, err := c.MultipartForm()

	if err != nil {
		return c.JSON(400, ErrorResponse{
			Error: "You must provide files.",
		})
	}

	files := form.File["files"]

	secretData := form.Value["secret-data"][0]
	var claimsMap jwt.MapClaims
	p, err := jwt.ParseWithClaims(secretData, &claimsMap, func(t *jwt.Token) (interface{}, error) {
		key, err := jwt.ParseRSAPublicKeyFromPEM(cert)
		if err != nil {
			return nil, err
		}

		return key, nil
	})

	if err != nil {
		println(err.Error())
		return err
	}

	if !p.Valid {
		return fmt.Errorf("message verification failed")
	}

	var tempFileNamesMapping []TempFileMapping

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

		newFile, err := os.CreateTemp(FILES_FOLDER, "*")

		if err != nil {
			return c.JSON(400, ErrorResponse{
				Error: "Unknown error.",
			})
		}

		tempFileNamesMapping = append(tempFileNamesMapping, TempFileMapping{
			OriginalFileName: file.Filename,
			NewFileName:      filepath.Base(newFile.Name()),
		})

		if err != nil {
			panic(err.Error())
		}

		_, err = io.Copy(newFile, dataStream)

		if err != nil {
			return c.JSON(400, ErrorResponse{
				Error: "Unknown error.",
			})
		}
	}

	println(tempFileNamesMapping)

	return c.JSON(200, tempFileNamesMapping)
}

func getFile(c echo.Context) error {
	fileName := c.Param("fileName")
	if fileName == "" {
		c.JSON(400, ErrorResponse{
			Error: "missing filename",
		})
	}
	println(os.Getwd())
	println(fileName)
	return c.File(path.Join(FILES_FOLDER, fileName))
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func authorizeUserWithNextAuthServer(c *echo.Context) bool {
	ct := *c

	client := &http.Client{}
	req, _ := http.NewRequest("GET", config.GlobalAppConfig.AuthServerUrl, nil)
	req.Header = ct.Request().Header

	cookies := ct.Request().Cookies()

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != 200 {
		return false
	}

	bodyContent, _ := io.ReadAll(resp.Body)

	if string(bodyContent) == "{}" {
		log.Printf("Unauthorized! Cookies: ")

		for _, cookie := range cookies {
			println(cookie.Name + ": " + cookie.Value)
		}

		return false
	}

	log.Println(string(bodyContent))

	return true
}

type TempFileMapping struct {
	OriginalFileName string `json:"originalFileName"`
	NewFileName      string `json:"newFileName"`
}
