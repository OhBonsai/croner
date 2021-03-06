package croner

import (
	"fmt"
	"gopkg.in/robfig/cron.v2"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const (
	IDLE = iota
	RUNNING
	FAIL
	STOP
)

type JobRunError struct {
	Message string
}

func (err JobRunError) Error() string {
	return err.Message
}

type WrappedJob struct {
	Id           int
	Name         string
	Inner        JobInf
	status       uint32
	running      sync.Mutex
	SuccessCount uint32
	TotalCount   uint32
	father       *CronManager
	Info         interface{}
	Next         time.Time
}

type JobRunReturnWithEid struct {
	JobRunReturn
	Eid int
}

type JobRunReturn struct {
	Value interface{}
	Error error
}

type JobInf interface {
	Run() JobRunReturn
}

func NewWrappedJob(job JobInf, r *CronManager) *WrappedJob {
	name := reflect.TypeOf(job).Name()
	return &WrappedJob{
		Inner:  job,
		father: r,
		Name:   name,
	}
}

func (j *WrappedJob) Status() string {
	switch atomic.LoadUint32(&j.status) {
	case RUNNING:
		return "RUNNING"
	case IDLE:
		return "IDLE"
	case STOP:
		return "STOP"
	default:
		return "FAIL"
	}
}

func (j *WrappedJob) Now() {
	defer func() {
		j.TotalCount += 1
		if err := recover(); err != nil {
			errString := fmt.Sprintf("WrappedJob-%d %s  execute fail. error is %s", j.Id, j.Name, err)
			println(errString)
			atomic.StoreUint32(&j.status, FAIL)
			if !j.father.ignorePanic {
				j.father.DisActive(j.Id)
			}
			j.father.jobReturnsWithEid <- JobRunReturnWithEid{
				JobRunReturn{nil, JobRunError{errString}},
				j.Id,
			}
		}
		return
	}()

	//print("Goroutine ID is ", goid.Get(), "\n")
	//print("Plan<", j.Name, "> ", time.Now().Minute(),":" ,time.Now().Second(), "   Wrapped JOB <", j,"> running\n")

	// 同一时间，该任务只会执行一次， 比如一个任务要执行1小时，周期设置5s。 那么不会有更多的协程出来多次执行该任务
	if j.father.onlyOne {
		j.running.Lock()
		defer j.running.Unlock()
	}

	if j.father.poolSize > 0 && permit != nil {
		permit <- struct{}{}
		defer func() { <-permit }()
	}

	atomic.StoreUint32(&(j.status), RUNNING)
	defer atomic.StoreUint32(&(j.status), IDLE)

	if j.father.timeInterrupt > 0 {
		t := time.NewTimer(time.Duration(j.father.timeInterrupt) * time.Second)
		var done = make(chan bool)
		defer close(done)
		go func() {
			j.execute()
			done <- true
		}()
		select {
		case <-t.C:
			panic(fmt.Sprint("Timeout ", j.father.timeInterrupt, "s"))
		case <-done:
			return
		}
	}

	j.execute()
}

func (j *WrappedJob) Run() {
	j.Now()
	return
}

func (j *WrappedJob) execute() {
	j.father.jobReturnsWithEid <- JobRunReturnWithEid{j.Inner.Run(), j.Id}
	j.SuccessCount += 1
	if j.father.onlyOne {
		// sometimes will fail.
		j.Next = j.father.MainCron.Entry(cron.EntryID(j.Id)).Next
	}else{
		j.Next = j.father.MainCron.Entry(cron.EntryID(j.Id)).Next
	}
}
