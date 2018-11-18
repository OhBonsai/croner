package croner


var DefaultManager *CronManager
var permit chan struct{}


func SetDefaultManager(manager *CronManager){
	DefaultManager = manager
}

func init() {
	m := NewCronManager(CronManagerConfig{true, false, 0, 60})
	SetDefaultManager(m)
}