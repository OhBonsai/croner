package restcron

import (
	"gopkg.in/robfig/cron.v2"
)

type CronManager struct {
	JobMap        map[int] *WrappedJob
	mainCron      *cron.Cron
	jobReturns    chan JobRunReturn
	stop          chan struct{}
	ignorePanic   bool
	onlyOne       bool
	poolSize      uint
	timeInterrupt uint
	running       bool
}


func NewCronManager(c CronManagerConfig) *CronManager {
	if c.PoolSize > 0 {
		permit = make(chan struct{}, c.PoolSize)
	}
	return &CronManager{
		make(map[int] *WrappedJob),
		cron.New(),
		make(chan JobRunReturn, 10),
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
		panic("Can't set config when runner running")
	}

	if r.poolSize != cfg.PoolSize && cfg.PoolSize > 0 {
		permit = make(chan struct{}, cfg.PoolSize)
	}

	r.ignorePanic = cfg.IgnorePanic
	r.onlyOne = cfg.OnlyOne
	r.poolSize = cfg.PoolSize
	r.timeInterrupt = cfg.TimeInterrupt
}

func (r *CronManager) Add(spec string, j JobInf) (int, error){
	schedule, err := cron.Parse(spec)
	if err != nil {
		return -1, err
	}
	wrappedJob := NewWrappedJob(j, r)
	entryId := int(r.mainCron.Schedule(schedule, wrappedJob))
	r.JobMap[entryId] = wrappedJob
	wrappedJob.Id = entryId
	return entryId, nil
}

func (r *CronManager) Remove(id int) {
	r.mainCron.Remove(cron.EntryID(id))
	delete(r.JobMap, id)
}

func (r *CronManager) DisActive(id int) {
	r.mainCron.Remove(cron.EntryID(id))
	r.JobMap[id].status = STOP
}

func (r *CronManager) RemoveAll() {
	// It's safe when deleting key in loop
	for id, _ := range r.JobMap {
		r.Remove(id)
	}
}

func (r *CronManager) Start() {
	r.mainCron.Start()
	r.run()
	r.running = true
}

func (r *CronManager) Stop() {
	r.mainCron.Stop()
	r.stop <- struct{}{}
	r.running = false
}


func (r *CronManager) run() {
	go func(){
		for {
			select {
			case value := <-r.jobReturns:
				jobReturnHooks.Run(&value)
			case <-r.stop:
				return
			}
		}
	}()
}





