package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/DeanThompson/ginpprof"
	"fmt"
	"os"
	"time"
	"github.com/OhBonsai/croner"
	"net/http"
	"sort"
)

var manager = croner.NewCronManager(croner.CronManagerConfig{
	true, false, 0, 0,
})

var ch = make(chan string, 10)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1,
	WriteBufferSize: 10240,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type JobS struct {
	Duration int    `json:"duration"`
	Who      string `json:"who"`
	What     string `json:"what"`
}

func (j JobS) Run() croner.JobRunReturn {
	return croner.JobRunReturn{
		Value: fmt.Sprintf("[%s] %s: %s", time.Now().Format(time.RFC850), j.Who, j.What),
	}
}

func CreateJob(c *gin.Context) {
	var curJob JobS
	err := c.BindJSON(&curJob)

	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	// you can put some info into job
	_, err = manager.Add(fmt.Sprintf("@every %ds", curJob.Duration), curJob, nil)
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
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
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	SuccessAndTotal string    `json:"success"`
	Next            time.Time `json:"next"`
	Eid             int       `json:"-"`
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

		for _, v := range manager.JobMap {
			innerJob := v.Inner.(JobS)
			returnResponse = append(returnResponse, StatusResp{
				Name:            innerJob.Who,
				Status:          v.Status(),
				SuccessAndTotal: fmt.Sprintf("%d/%d", v.SuccessCount, v.TotalCount),
				Next:            v.Next,
				Eid:             int(v.Id),
			})
		}
		sort.Slice(returnResponse, func(i, j int) bool {
			return returnResponse[i].Eid < returnResponse[j].Eid
		})
		conn.WriteJSON(returnResponse)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	// push job run return into channel
	croner.OnJobReturn(func(runReturn *croner.JobRunReturn) {
		say := runReturn.Value.(string)
		ch <- say
	})
	croner.SetDefaultManager(manager)
	manager.Start()

	r := gin.Default()
	r.LoadHTMLGlob("example/*.html")
	ginpprof.Wrapper(r)
	r.POST("/job", CreateJob)
	r.GET("/echo", Echo)
	r.GET("/status", Status)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.Run(":8000")
}
