package main

import (
	"context"
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
	urlPrefix := os.Getenv("URL_BASE")

	r := gin.Default()
	db := database.NewPostgresDB()
	defer db.Close()
	rdb := database.NewRedisDB()
	defer rdb.Close()

	repo := repository.NewLinkRepository(db, rdb)
	linkService := service.NewLinkService(repo)
	linkHandler := handler.NewLinkHandler(linkService)

	go service.StartFlusher(context.Background(), rdb, db)

	go bot.StartBot(linkService, urlPrefix)

	r.GET("/links/:short", linkHandler.GetLink)
	r.POST("/links", linkHandler.CreateLink)
	r.GET("/links", linkHandler.GetLinks)
	r.DELETE("/links/:short", linkHandler.DeleteLink)
	r.GET("/:short", linkHandler.Redirect)
	r.GET("/clicks/:short", linkHandler.GetClicks)

	r.Run(":8080")
}
