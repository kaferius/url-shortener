package main

import (
	"log"
	"os"
	"url-shortener/database"
	"url-shortener/handler"
	"url-shortener/handler/telegram"
	"url-shortener/repository"
	"url-shortener/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	bot, err := telegram.NewBotHandler(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal("creating bot error")
	}

	r := gin.Default()
	db := database.NewPostgresDB()
	defer db.Close()
	rdb := database.NewRedisDB()
	defer rdb.Close()

	repo := repository.NewLinkRepository(db, rdb)
	service := service.NewLinkService(repo)
	handler := handler.NewLinkHandler(service)

	go bot.StartBot(service)

	r.GET("/links/:short", handler.GetLink)
	r.POST("/links", handler.CreateLink)
	r.GET("/links", handler.GetLinks)
	r.DELETE("/links/:short", handler.DeleteLink)
	r.GET("/:short", handler.Redirect)

	r.Run(":8080")
}
