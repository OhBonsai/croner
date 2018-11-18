package restcron


var DefaultManager *CronManager
var permit chan struct{}


func SetDefaultRunner(runner *CronManager){
	DefaultManager = runner
}

func init() {
	DefaultManager = NewCronManager(CronManagerConfig{true, false, 0, 60})
}