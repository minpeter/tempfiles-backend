package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/minpeter/tempfiles-backend/database"
	"github.com/minpeter/tempfiles-backend/file"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/joho/godotenv/autoload"

	jwtWare "github.com/gofiber/jwt/v3"
)

type LoginRequest struct {
	Email    string
	Password string
}

func main() {

	VER := "1.1.6"
	app := fiber.New(fiber.Config{
		AppName:   "tempfiles-backend",
		BodyLimit: int(math.Pow(1024, 3)), // 1 == 1byte
	})

	app.Use(
		cache.New(cache.Config{
			StoreResponseHeaders: true,
			Next: func(c *fiber.Ctx) bool {
				return c.Route().Path != "/dl/:filename"
			},
		}),
		cors.New(cors.Config{
			AllowOrigins: "*",
			AllowHeaders: "Origin, Content-Type, Accept",
			AllowMethods: "GET, POST, DELETE",
		}))

	var err error

	file.MinioClient, err = file.Connection()
	if err != nil {
		log.Fatalf("minio connection error: %v", err)
	}

	database.Engine, err = database.CreateDBEngine()
	if err != nil {
		log.Fatalf("failed to create db engine: %v", err)
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":    "api is working normally :)",
			"apiVersion": VER,
		})
	})

	app.Get("/info", func(c *fiber.Ctx) error {
		apiName := c.Query("api", "")
		switch apiName {
		case "upload":
			return c.JSON(fiber.Map{
				"apiName": "/upload",
				"method":  "POST",
				"desc":    "특정 파일을 서버에 업로드합니다.",
				"command": "curl -X POST -F 'file=@[filepath or filename]' https://tfb.minpeter.cf/upload",
			})
		case "list":
			return c.JSON(fiber.Map{
				"apiName": "/list",
				"method":  "GET",
				"desc":    "서버에 존재하는 파일 리스트를 반환합니다.",
				"command": "curl https://tfb.minpeter.cf/list",
			})
		case "del":
			return c.JSON(fiber.Map{
				"apiName": "/del/[filename]",
				"method":  "DELETE",
				"desc":    "서버에 존재하는 특정 파일을 삭제합니다.",
				"command": "curl -X DELETE https://tfb.minpeter.cf/del/[filename]",
			})
		case "dl":
			return c.JSON(fiber.Map{
				"apiName": "/dl/[filename]",
				"method":  "GET",
				"desc":    "서버에 존재하는 특정 파일을 다운로드 합니다.",
				"command": "curl -O https://tfb.minpeter.cf/dl/[filename]",
			})
		default:
			backendUrl := os.Getenv("BACKEND_BASEURL")
			return c.JSON([]fiber.Map{
				{
					"apiUrl":     backendUrl + "/upload",
					"apiHandler": "upload",
				},
				{
					"apiUrl":     backendUrl + "/list",
					"apiHandler": "list",
				},
				{
					"apiUrl":     backendUrl + "/del/[filename]",
					"apiHandler": "del",
				},
				{
					"apiUrl":     backendUrl + "/dl/[filename]",
					"apiHandler": "dl",
				},
			})
		}
	})

	app.Get("/list", file.ListHandler)

	app.Post("/upload", file.UploadHandler)
	app.Get("/checkpw/:filename", file.CheckPasswordHandler)

	app.Use(jwtWare.New(jwtWare.Config{
		SigningKey:  []byte(os.Getenv("JWT_SECRET")),
		TokenLookup: "query:token",
		Filter: func(c *fiber.Ctx) bool {

			fileName := strings.Split(strings.Split(c.OriginalURL(), "/")[2], "?")[0]

			fileRow := new(database.FileRow)
			has, err := database.Engine.Where("file_name = ?", fileName).Desc("id").Get(fileRow)
			if err != nil {
				return false
			}
			if !has {
				return false
			}
			return !fileRow.Encrypto
		},
	}))

	app.Get("/dl/:filename", file.OldDownloadHandler)
	app.Delete("/del/:filename", file.DeleteHandler)

	log.Fatal(app.Listen(fmt.Sprintf(":%s", os.Getenv("BACKEND_PORT"))))

}
