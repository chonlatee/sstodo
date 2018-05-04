package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	_ "github.com/line/line-bot-sdk-go/linebot"
)

func main() {

	port := os.Getenv("PORT")
	selfHost := "https://ssbotline.herokuapp.com"

	// bot_channel_id := os.Getenv("bot_channel_id")
	// bot_channel_secret := os.Getenv("bot_channel_secret")
	login_channel_id := os.Getenv("login_channel_id")
	// login_channel_secret := os.Getenv("login_channel_secret")

	// bot, err := linebot.New(
	// 	os.Getenv("channel_token"),
	// 	os.Getenv("channel_secret"),
	// )

	// fmt.Println(os.Getenv("channel_secret"))
	// fmt.Println(os.Getenv("channel_token"))

	// if err != nil {
	// 	log.Fatalln(err)
	// }

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

	r.Static("/assets", "./assets")

	// r.POST("/callback", func(c *gin.Context) {
	// 	events, err := bot.ParseRequest(c.Request)
	// 	if err != nil {
	// 		if err == linebot.ErrInvalidSignature {
	// 			c.JSON(400, gin.H{
	// 				"err": "Invalid signature error",
	// 			})
	// 		} else {
	// 			c.JSON(500, gin.H{
	// 				"err": "server error",
	// 			})
	// 		}
	// 		return
	// 	}

	// 	for _, event := range events {
	// 		log.Printf("Got event %v", event)
	// 		fmt.Println()
	// 		if event.Type == linebot.EventTypeMessage {
	// 			switch message := event.Message.(type) {
	// 			case *linebot.TextMessage:
	// 				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
	// 					log.Print(err)
	// 				}
	// 			}
	// 		}
	// 	}
	// })

	r.GET("/", func(c *gin.Context) {

		state := uniuri.New()

		r.LoadHTMLGlob("templates/*")
		loginCallbackURI := selfHost + "/logincallback"
		lineloginURL := "https://access.line.me/dialog/oauth/weblogin" +
			"?response_type=code&client_id=" + login_channel_id + "&redirect_uri=" + url.QueryEscape(loginCallbackURI) + "&state=" + state

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":        "Simple Stupid Todo",
			"lineloginURI": lineloginURL,
		})
	})

	r.GET("/logincallback", func(c *gin.Context) {
		u, err := url.Parse(c.Request.URL.String())
		if err != nil {
			fmt.Println("can't parse query")
		}
		fmt.Printf("%v", u)
		// if has code query
		// use code for get access token
		// use access token for get user email
	})

	// router for manage todo
	// r.POST("/todo", func(c *gin.Context) {

	// })

	// r.GET("/todo", func(c *gin.Context) {

	// })

	// r.GET("/todo/:id", func(c *gin.Context) {

	// })

	r.Run(":" + port)
}

// create function for split message
// create database database for save todo
// use mongo for store data
// create template for edit todo
func splitMessage(msg string) {
	msgSplit := strings.Split(msg, ":")

	if len(msgSplit) < 2 {
		log.Fatalln("invalid format")
		return
	}

	if len(msgSplit) > 4 {
		log.Fatalln("invalid format")
		return
	}

	if len(msgSplit) == 2 {
		fmt.Println("set detault time")
	}

	if len(msgSplit) == 3 {
		fmt.Println("set task todo")
	}
}
