# croner

Manage crontab in golang， **Star is the best praise and gift...**


## Function
In development, we often need to set time tasks. [Golang Cron](https://github.com/robfig/cron/tree/v2) provides some
good basic schedule functions. But in some complex scenarios. It's not enough. Like:

- Any task could fail. If the task fails. try again or stop this task?
- The user set `@every 5s` execution. But It need 10 sec to finish this task. How to ensure a task is performed only once per unit time.
- I wanna watch all tasks' status. Is it running? What's the next execution time.
- I wanna know how many times the task executed and how many times it was successful
- I wanna limit the number of concurrent tasks
- I can return some results when the task is complete. And I can add some hook functions to handle those results

View the [test](https://github.com/OhBonsai/croner/blob/master/manager_test.go) to know more...


## Example

Run example
```
go get -u "github.com/OhBonsai/croner"
go get -u "github.com/gin-gonic/gin"
go get -u "github.com/gorilla/websocket"
cd $GOPATH/src/github.com/OhBonsai/croner/example

go run server.go
# Open localhost:8000
```

![image](https://upload-images.jianshu.io/upload_images/3981759-cf668d205086d9bc.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240
)

1. define some time task
2. view all task status
3. each task will send data to terminal periodically




## 中文

**360度托马斯回旋平沙落雁五体投地雪地跪求点赞**

在开发中，经常遇到一些需要定时任务的场景。各个语言都有定时语言的库，[Golang Cron](https://github.com/robfig/cron/tree/v2) 提供了Crontab Golang语言版本。这个库非常不错，提供最基本的定时任务编排的功能。但是一些复杂需求无法满足，比如
- 任何定时任务都有可能失败，失败了就panic了，这样非常不友好。最起码能够让我控制，失败是重试还是停止
- 某些任务执行周期要10s, 而用户设置的5s一执行，我能不能保证任何时间这个任务只执行一次
- 我想实时的看到任务的状态，比如是不是在运行？下次运行时间？上次运行时间？
- 我想看到任务执行了多少次，成功了多少次
- 我想要限制最大任务数量，比如超过10个任务在执行，不运行新的任务执行
- 任务执行完了可以告诉我逻辑上有错误，还是有结果。我还可以加上一些钩子函数来处理任务执行的结果

[详细请查看博客]()





