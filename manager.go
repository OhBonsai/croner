package croner

import (
	"gopkg.in/robfig/cron.v2"
	"time"
)

type CronManager struct {
	JobMap            map[int] *WrappedJob
	MainCron          *cron.Cron
	jobReturnsWithEid chan JobRunReturnWithEid
	stop              chan struct{}
	ignorePanic       bool
	onlyOne           bool
	poolSize          uint
	timeInterrupt     uint
	running           bool
}


func NewCronManager(c CronManagerConfig) *CronManager {
	if c.PoolSize > 0 {
		permit = make(chan struct{}, c.PoolSize)
	}
	return &CronManager{
		make(map[int] *WrappedJob),
		cron.New(),
		make(chan JobRunReturnWithEid, 10),
		make(chan struct{}),
		c.IgnorePanic,
		c.OnlyOne,
		c.PoolSize,
		c.TimeInterrupt,
		false,
	}
}


type CronManagerConfig struct {
	IgnorePanic   bool
	OnlyOne       bool
	PoolSize      uint
	TimeInterrupt uint
}


func (r *CronManager) SetConfig(cfg CronManagerConfig) {
	if r.running {
		panic("Can't set config when manager running")
	}

	if r.poolSize != cfg.PoolSize && cfg.PoolSize > 0 {
		permit = make(chan struct{}, cfg.PoolSize)
	}

	r.ignorePanic = cfg.IgnorePanic
	r.onlyOne = cfg.OnlyOne
	r.poolSize = cfg.PoolSize
	r.timeInterrupt = cfg.TimeInterrupt
}

func (r *CronManager) Add(spec string, j JobInf, info interface{}) (int, error){
	schedule, err := cron.Parse(spec)
	if err != nil {
		return -1, err
	}
	wrappedJob := NewWrappedJob(j, r)
	entryId := int(r.MainCron.Schedule(schedule, wrappedJob))
	r.JobMap[entryId] = wrappedJob
	wrappedJob.Id = entryId
	wrappedJob.Info = info
	return entryId, nil
}

func (r *CronManager) Remove(id int) {
	r.MainCron.Remove(cron.EntryID(id))
	delete(r.JobMap, id)
}

func (r *CronManager) DisActive(id int) {
	r.MainCron.Remove(cron.EntryID(id))
	r.JobMap[id].status = STOP
}

func (r *CronManager) RemoveAll() {
	// It's safe when deleting key in loop
	for id := range r.JobMap {
		r.Remove(id)
	}
}

func (r *CronManager) Start() {
	if !r.running{
		r.MainCron.Start()
		r.run()
		r.running = true
	}
}

func (r *CronManager) Stop() {
	r.MainCron.Stop()
	r.stop <- struct{}{}
	r.running = false
}

func (r *CronManager) Job(id int) (*WrappedJob, bool) {
	v, ok := r.JobMap[id]
	return v, ok
}

func (r *CronManager) run() {
	go func(){
		for {
			select {
			case value := <-r.jobReturnsWithEid:
				jobReturnHooks.Run(&value)
			case <-r.stop:
				return
			}
		}
	}()
}

func Validate(spec string) bool{
	_, err := cron.Parse(spec)
	return err == nil
}

func Next(spec string) (time.Time, error) {
	schema, err := cron.Parse(spec)
	if err != nil {
		return time.Time{}, err
	}

	return schema.Next(time.Now()), nil
}



