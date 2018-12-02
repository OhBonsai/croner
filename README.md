# croner
对[CRON.V2](https://github.com/robfig/cron/tree/v2)进行管理。
1. 当一个任务PANIC了，提供参数`IGNORE_PANIC`选择忽略错误或者不再执行
2. 提供任务的状态(`IDLE`, `FAIL`, `STOP`, `RUNNING`)
3. 记录任务执行成功次数／总次数
4. 提供参数`ONLY_ONE`支持 只有任务前一次执行完成，才执行。 比如一个任务执行5分钟，周期为一分钟，五分钟只会执行一次
5. 提供最大任务数量参数`POOL_SIZE`
6. 提供接口查看所有任务的状态


## Example
例子为`gin`,`websocket`的在线聊天室。通过定义`schedule`(支持crontab和rubyclock)。让服务器周期发送消息到客户端。
这是一个没有实际用处的例子，但很好的演示了在WEB场景下如何`Manage`一个后台的定时任务。

### 启动
```
go get -u "github.com/OhBonsai/croner"
go get -u "github.com/gin-gonic/gin"
go get -u "github.com/gorilla/websocket"
cd $GOPATH/src/github.com/OhBonsai/croner/example

go run server.go 
# Open localhost:8000
```

