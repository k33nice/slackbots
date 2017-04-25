package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/k33nice/slackbots/dobby/service"
)

// Command - interface for system command.
type Command interface {
	Run(cmd string, wg *sync.WaitGroup)
	GetResult() *service.Result
}

// API - object that handle api http-server and execute commands.
type API struct {
	Log     *log.Logger
	Port    int
	Token   string
	Service Command
}

// Run - api method that start http-server, that treat slack "Slash Commands".
func (a *API) Run() {
	gin.SetMode(gin.ReleaseMode)

	api := gin.New()
	api.Use(a.logger())
	api.Use(gin.Recovery())
	api.Use(a.auth())

	wg := new(sync.WaitGroup)

	api.POST("/run", func(c *gin.Context) {
		text := c.PostForm("text")
		a.Log.Printf("run handler recieve: %s command", text)
		wg.Add(1)
		go a.Service.Run(text, wg)
		wg.Wait()
		res := a.Service.GetResult()
		a.Log.Println(res)
		var result string

		if res.IsOk {
			result = "DONE!"
		} else {
			result = fmt.Sprintf("Cannot execute your comand, error: %s", res.Stdout)
		}

		c.JSON(200, gin.H{
			"text": result,
		})
	})
	api.Run(strings.Join([]string{":", strconv.Itoa(a.Port)}, ""))
}

func (a *API) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.PostForm("token")

		if token != a.Token {
			respondWithError(http.StatusUnauthorized, "UNAUTHORIZED", c)
			return
		}

		c.Next()
	}
}

func (a *API) logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		a.Log.Println(status)
	}
}

func respondWithError(code int, message string, c *gin.Context) {
	resp := map[string]string{"error": message}

	c.JSON(code, resp)
	c.Abort()
}
