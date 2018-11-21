package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"fmt"
	"os"
	"time"
	"github.com/OhBonsai/croner"
	"net/http"
)

var manager = croner.NewCronManager(croner.CronManagerConfig{
	true, false, 0, 0,
})

var ch = make(chan string, 10)
var upgrader = websocket.Upgrader{
	ReadBufferSize: 1,
	WriteBufferSize: 10240,
	CheckOrigin: func(r *http.Request) bool {return true},
}

type JobS struct {
	Duration int    `json:"Duration"`
	Who      string `json:"Who"`
	What     string `json:"What"`
}


func (j JobS) Run() croner.JobRunReturn{
	return croner.JobRunReturn{
		Value: fmt.Sprintf("[%s] %s: %s",time.Now().Format(time.RFC850), j.Who, j.What),
	}
}


func CreateJob(c *gin.Context) {
	var curJob JobS
	err := c.BindJSON(&curJob)

	if err != nil {
		c.JSON(400, "Bad Request")
	}

	// you can put some info into job
	manager.Add(fmt.Sprintf("@every %ds", curJob.Duration), curJob, nil)
	c.JSON(200, "success")
}


func Echo(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Sprintf("Failed to set websocket upgrader: %v", err)
		time.Sleep(2000)
		os.Exit(1)
	}

	// push data to client
	for {
		select {
			case msg := <-ch:
				conn.WriteMessage(websocket.TextMessage, []byte(msg))
			default:
				continue
		}
	}
}

type StatusResp struct {
	Name string `json:"name"`
	Status string `json:"status"`
	SuccessAndTotal string `json:"success"`
}

// all job status
func Status(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Sprintf("Failed to set websocket upgrader: %v", err)
		time.Sleep(2000)
		os.Exit(1)
	}

	for {
		returnResponse := []StatusResp{}

		for _, v := range manager.JobMap{
			innerJob := v.Inner.(JobS)
			returnResponse = append(returnResponse, StatusResp{
				Name: innerJob.Who,
				Status: v.Status(),
				SuccessAndTotal: fmt.Sprintf("%d/%d", v.SuccessCount, v.TotalCount),
			})
		}
		conn.WriteJSON(returnResponse)
		time.Sleep(1 * time.Second)
	}
}


func main(){
	// push job run return into channel
	croner.OnJobReturn(func(runReturn *croner.JobRunReturn) {
		say := runReturn.Value.(string)
		ch <- say
	})
	croner.SetDefaultManager(manager)
	manager.Start()

	r := gin.Default()
	r.LoadHTMLGlob("example/*.html")
	r.POST("/job", CreateJob)
	r.GET("/echo", Echo)
	r.GET("/status", Status)
	r.GET("/", func(c *gin.Context){
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.Run(":8000")
}
