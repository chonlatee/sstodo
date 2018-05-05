package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/dchest/uniuri"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

// AcccessTokenResult ...
type AcccessTokenResult struct {
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpireIn     int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func main() {

	port := os.Getenv("PORT")
	selfHost := "https://ssbotline.herokuapp.com"

	botChannelID := os.Getenv("bot_channel_id")
	botChannelSecret := os.Getenv("bot_channel_secret")
	loginChannelID := os.Getenv("login_channel_id")
	loginChannelSecret := os.Getenv("login_channel_secret")

	bot, err := linebot.New(
		botChannelID,
		botChannelSecret,
	)

	if err != nil {
		log.Fatalln(err)
	}

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("sstodoSession", store))
	r.Static("/assets", "./assets")
	r.Use(gin.Logger())

	r.POST("/callback", func(c *gin.Context) {
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
			log.Printf("Got event %v", event)
			fmt.Println()
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

	r.GET("/", func(c *gin.Context) {
		// get access token from session
		// if no access token redirect to line login url
		// if access token exist and not expire
		// if access token expire use refresh token for get new access token
		// but in line docs i did't see how to revoke access token with refresh token

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":     "Simple Stupid Todo",
			"linelogin": "/linelogin",
		})
	})

	r.GET("/linelogin", func(c *gin.Context) {
		state := uniuri.New()
		session := sessions.Default(c)
		session.Set("state", state)
		r.LoadHTMLGlob("templates/*")
		loginCallbackURI := selfHost + "/logincallback"
		lineloginURL := "https://access.line.me/dialog/oauth/weblogin" +
			"?response_type=code&client_id=" + loginChannelID + "&redirect_uri=" + url.QueryEscape(loginCallbackURI) + "&state=" + state
		session.Save()
		c.Redirect(301, lineloginURL)
	})

	r.GET("/logincallback", func(c *gin.Context) {

		code := c.Query("code")
		state := c.Query("state")
		errDesc := c.Query("error_description")

		session := sessions.Default(c)
		stateSes := session.Get("state")

		if state != stateSes {
			c.Redirect(301, "/loginError?err_desc=state_not_match")
		}

		if len(errDesc) != 0 {
			c.Redirect(301, "/loginError?err_desc="+errDesc)
		}

		log.Println("Get access token")

		// Get access token
		accessTokenURL := "https://api.line.me/"
		resource := "v2/oauth/accessToken"
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("client_id", loginChannelID)
		data.Set("client_secret", loginChannelSecret)
		data.Set("code", code)
		data.Set("redirect_uri", selfHost+"/logincallback")

		u, _ := url.ParseRequestURI(accessTokenURL)
		u.Path = resource
		u.RawQuery = data.Encode()
		urlStr := fmt.Sprintf("%v", u)

		client := &http.Client{}
		req, err := http.NewRequest("POST", urlStr, nil)
		if err != nil {
			log.Fatalln("parse req error", err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln("Get access token error", err)
		}
		defer resp.Body.Close()

		tokenResult := AcccessTokenResult{}
		json.NewDecoder(resp.Body).Decode(&tokenResult)

		if len(tokenResult.AccessToken) != 0 {
			log.Println("redirect to dashboard")
			session.Set("accToken", tokenResult.AccessToken)
			fmt.Println(tokenResult.Scope)
			fmt.Println(tokenResult.AccessToken)
			fmt.Println(tokenResult.TokenType)
			fmt.Println(tokenResult.ExpireIn)
			fmt.Println(tokenResult.RefreshToken)
			session.Save()
			c.HTML(http.StatusOK, "redirect.tmpl", gin.H{
				"msg":  "login success go to dashboard",
				"link": "/dashboard",
			})
		}
	})

	r.GET("/loginError", func(c *gin.Context) {
		errDesc := c.Query("err_desc")
		c.JSON(200, gin.H{
			"err": errDesc,
		})
	})

	r.GET("/dashboard", func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("accToken")
		log.Printf("token %v\n", token)
		if token == nil {
			c.Redirect(301, "/")
		}
		c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
			"title": "Simple Stupid Todo",
		})
	})

	r.GET("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Delete("accToken")
		session.Save()
		c.Redirect(301, "/")
	})

	// 	// router for manage todo
	// 	// r.POST("/todo", func(c *gin.Context) {

	// 	// })

	// 	// r.GET("/todo", func(c *gin.Context) {

	// 	// })

	// 	// r.GET("/todo/:id", func(c *gin.Context) {

	// 	// })

	r.Run(":" + port)
}

// // create function for split message
// // create database database for save todo
// // use mongo for store data
// // create template for edit todo
// func splitMessage(msg string) {
// 	msgSplit := strings.Split(msg, ":")

// 	if len(msgSplit) < 2 {
// 		log.Fatalln("invalid format")
// 		return
// 	}

// 	if len(msgSplit) > 4 {
// 		log.Fatalln("invalid format")
// 		return
// 	}

// 	if len(msgSplit) == 2 {
// 		fmt.Println("set detault time")
// 	}

// 	if len(msgSplit) == 3 {
// 		fmt.Println("set task todo")
// 	}
// }
