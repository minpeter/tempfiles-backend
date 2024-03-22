package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/robfig/cron"
	"github.com/rs/cors"

	_ "github.com/joho/godotenv/autoload"
	controller "github.com/tempfiles-Team/tempfiles-backend/controllers"
	"github.com/tempfiles-Team/tempfiles-backend/database"
	"github.com/tempfiles-Team/tempfiles-backend/utils"
)

func main() {

	if os.Getenv("BACKEND_PORT") == "" {
		os.Setenv("BACKEND_PORT", "5000")
	}

	port, _ := strconv.Atoi(os.Getenv("BACKEND_PORT"))
	s := fuego.NewServer(
		// string to int
		fuego.WithPort(port),
		fuego.WithCorsMiddleware(cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete},
			AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "X-Download-Limit", "X-Time-Limit", "X-Hidden"},
		}).Handler),
	)

	// app.Use(limits.RequestSizeLimiter(int64(math.Pow(1024, 3)))) // 1 == 1byte, = 1GB

	terminator := cron.New()

	terminator.AddFunc("1 */5 * * *", func() {
		log.Println("⏲️  Check for expired files", time.Now().Format("2006-01-02 15:04:05"))
		var files []database.FileTracking
		if err := database.Engine.Where("expire_time < ? and is_deleted = ?", time.Now(), false).Find(&files); err != nil {
			log.Println("cron db query error", err.Error())
		}
		for _, file := range files {
			log.Printf("🗑️  Set this folder for deletion: %s \n", file.FolderId)
			file.IsDeleted = true
			if _, err := database.Engine.ID(file.Id).Cols("Is_deleted").Update(&file); err != nil {
				log.Printf("cron db update error, file: %s, error: %s\n", file.FolderId, err.Error())
			}
		}
	})

	terminator.AddFunc("1 */20 * * *", func() {
		log.Println("⏲️  Check which files need to be deleted", time.Now().Format("2006-01-02 15:04:05"))
		var files []database.FileTracking
		if err := database.Engine.Where("is_deleted = ?", true).Find(&files); err != nil {
			log.Println("file list error: ", err.Error())
		}
		for _, file := range files {
			log.Printf("🗑️  Delete this folder: %s\n", file.FolderId)
			if err := os.RemoveAll("./tmp/" + file.FolderId); err != nil {
				log.Println("delete file error: ", err.Error())
			}
			if _, err := database.Engine.Delete(&file); err != nil {
				log.Println("delete file error: ", err.Error())
			}
		}
	})

	terminator.Start()

	var err error

	if utils.CheckTmpFolder() != nil {
		log.Fatalf("tmp folder error: %v", err)
	}

	if database.CreateDBEngine() != nil {
		log.Fatalf("failed to create db engine: %v", err)
	}

	fuego.Get(s, "/", func(c fuego.ContextNoBody) (string, error) {
		return "TEMPFILES API WORKING 🚀\nIf you want to use the API, go to '/swagger'", nil
	})

	controller.FilesRessources{}.Routes(s)

	s.Run()
	terminator.Stop()
}
