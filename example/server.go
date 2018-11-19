package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"fmt"
	"os"
)

var ch = make(chan string, 10)
var upgrader = websocket.Upgrader{
	ReadBufferSize: 1,
	WriteBufferSize: 10240,
}

type CreateJobRequestJson struct {
	duration int
	who string
	what string
}

func CreateJob(c *gin.Context) {

}

func Echo(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Sprintf("Failed to set websocket upgrader: %v", err)
		os.Exit(1)
	}

	for {
		select {
			case msg := <-ch:
				conn.WriteMessage(websocket.TextMessage, []byte(msg))
			default:
				continue
		}
	}
}

func main(){
	r := gin.Default()
	r.POST("/job", CreateJob)
	r.GET("/ws", Echo)
	r.Run(":80")
}
