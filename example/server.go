package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"fmt"
	"os"
	"time"
	"github.com/OhBonsai/croner"
)

var ch = make(chan string, 10)
var upgrader = websocket.Upgrader{
	ReadBufferSize: 1,
	WriteBufferSize: 10240,
}

type JobS struct {
	duration int
	who string
	what string
}

func (j JobS) Run() croner.JobRunReturn{
	return croner.JobRunReturn{
		Value: fmt.Sprintf("[%s] %s: %s",time.Now().Format(time.RFC850), j.who, j.what),
	}
}



func CreateJob(c *gin.Context) {
	var curJob JobS
	err := c.BindJSON(&curJob)

	if err != nil {
		c.JSON(400, "Bad Request")
	}


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

func Status(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Sprintf("Failed to set websocket upgrader: %v", err)
		os.Exit(1)
	}

	for {
		time.Sleep(1 * time.Second)
		conn.WriteJSON([]int{})
	}
}


func main(){
	croner.OnJobReturn(func(runReturn *croner.JobRunReturn) {
		say := runReturn.Value.(string)
		ch <- say
	})

	r := gin.Default()
	r.POST("/job", CreateJob)
	r.GET("/ws", Echo)
	r.GET("/status", Status)
	r.Run(":80")
}
