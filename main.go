package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/chonlatee/ssbot/todo"
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

// Profile ...
type Profile struct {
	UserID string `json:"userId"`
}

// Cors ...
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

func main() {

	port := os.Getenv("PORT")
	selfHost := os.Getenv("self_host")
	if selfHost == "" {
		selfHost = "https://ssbotline.herokuapp.com"
	}

	botChannelID := os.Getenv("bot_channel_id")
	botChannelSecret := os.Getenv("bot_channel_secret")
	loginChannelID := os.Getenv("login_channel_id")
	loginChannelSecret := os.Getenv("login_channel_secret")

	bot, err := linebot.New(
		botChannelSecret,
		botChannelID,
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
	r.Use(Cors())
	r.Static("/assets", "./assets")

	r.POST("/callback", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				log.Println("400")
				log.Println(err)
				c.JSON(400, gin.H{
					"err": "Invalid signature error",
				})
			} else {
				log.Println("500")
				log.Println(err)
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
					var replyMsg string
					if strings.ToLower(message.Text) == "edit" {
						replyMsg = selfHost
					} else {
						err := todo.Save(event.Source.UserID, message.Text)
						if err != nil {
							replyMsg = err.Error()
						} else {
							replyMsg = "Save success"
						}
					}

					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMsg)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}

	})

	r.GET("/", func(c *gin.Context) {

		session := sessions.Default(c)
		uid := session.Get("uid")

		if uid.(string) == "" {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"title":     "Simple Stupid Todo",
				"linelogin": "/linelogin",
			})
		} else {
			c.Redirect(http.StatusFound, "/todos")
			return
		}

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
		c.Redirect(http.StatusFound, lineloginURL)
		return
	})

	r.GET("/logincallback", func(c *gin.Context) {

		code := c.Query("code")
		state := c.Query("state")
		errDesc := c.Query("error_description")

		session := sessions.Default(c)
		stateSes := session.Get("state")

		if state != stateSes {
			c.Redirect(http.StatusFound, "/loginError?err_desc=state_not_match")
		}

		if len(errDesc) != 0 {
			c.Redirect(http.StatusFound, "/loginError?err_desc="+errDesc)
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

			getuserIDURL := "https://api.line.me/v2/profile"
			client := &http.Client{}
			req, _ := http.NewRequest("GET", getuserIDURL, nil)
			req.Header.Set("Authorization", "Bearer "+tokenResult.AccessToken)
			res, err := client.Do(req)
			if err != nil {
				log.Fatalln("get user id error")
			}
			defer res.Body.Close()

			profile := Profile{}

			json.NewDecoder(res.Body).Decode(&profile)

			session.Set("uid", profile.UserID)
			session.Save()
			c.HTML(http.StatusOK, "redirect.tmpl", gin.H{
				"msg":  "login success go to dashboard",
				"link": "/todos",
			})
		}
	})

	r.GET("/loginError", func(c *gin.Context) {
		errDesc := c.Query("err_desc")
		c.JSON(200, gin.H{
			"err": errDesc,
		})
	})

	r.GET("/todos", func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get("uid")
		if uid.(string) == "" {
			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"title":     "Simple Stupid Todo",
				"linelogin": "/linelogin",
			})
		} else {
			c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
				"title": "Simple Stupid Todo",
				"todos": todo.GetByUserID(uid.(string)),
			})
		}

	})

	r.GET("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Delete("uid")
		session.Save()
		c.Redirect(http.StatusFound, "/")
	})

	// if we have a lot of user we have to use redis cronJob system

	// users := users.GetAll()
	// for _, u := range users {
	// 	go func(userID string) {
	// 	}(u.UserID)
	// }

	r.Run(":" + port)
}
