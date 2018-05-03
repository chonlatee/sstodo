package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {

	port := os.Getenv("PORT")
	bot, err := linebot.New(
		os.Getenv("channel_secret"),
		os.Getenv("channel_token"),
	)

	if err != nil {
		log.Fatalln(err)
	}

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	r := gin.Default()
	r.Use(gin.Logger())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	r.GET("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				c.JSON(400, gin.H{
					"err": "Invalid signature error",
				})
			} else {
				c.JSON(500, gin.H{
					"err": "server error",
				})
			}
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}

	})

	r.Run(":" + port)
}
