package croner

import (
	"fmt"
	"testing"
	"time"
)

var tmp = [5]string{}
var manager = NewCronManager(CronManagerConfig{true, false, 0, 0})

// reset tmp array items to null string
func resetTmp() {
	tmp = [5]string{}
}

// remove all job in manager
func resetRunner() {
	if manager.running {
		manager.RemoveAll()
		manager.Stop()
	}
}

// panicJob will panic when running
type PanicJob struct{}

func (j PanicJob) Run() JobRunReturn {
	fmt.Println("Start runnning panic Job: ", time.Now().Minute(), ":", time.Now().Second())
	panic("hello, i am a panic job")
}

// returnSoonJob will return soon....
type ReturnSoonJob struct{}

func (j ReturnSoonJob) Run() JobRunReturn {
	fmt.Println("Start runnning good Job: ", time.Now().Minute(), ":", time.Now().Second())
	return JobRunReturn{"Hello , I am a good job", nil}
}

// time5SecJob will return after 5 seconds
type Time3SecJob struct{}

func (j Time3SecJob) Run() JobRunReturn {
	fmt.Println("Start runnning 3 sec Job: ", time.Now().Minute(), ":", time.Now().Second())
	time.Sleep(3 * time.Second)
	fmt.Println("End runnning 3 sec Job: ", time.Now().Minute(), ":", time.Now().Second())
	return JobRunReturn{"Hello , I am a timeout job", nil}
}

// hook function when job return. push value in tmp array
func hookAppendResultToTmp(runReturn *JobRunReturnWithEid) {
	fmt.Println("Hook running: ", time.Now().Minute(), ":", time.Now().Second())
	for i, v := range tmp {
		if v == "" {
			tmpStr := fmt.Sprintf("%v", runReturn.Value)
			tmp[i] = tmpStr
			break
		}
	}

}

// Simple test
func TestRunning(t *testing.T) {
	resetRunner()
	resetTmp()
	if len(jobReturnHooks) == 0 {
		OnJobReturn(hookAppendResultToTmp)
	}
	manager.Start()
	fmt.Println("Manager start running: ", time.Now().Minute(), ":", time.Now().Second())

	entryId, _ := manager.Add("@every 2s", ReturnSoonJob{}, nil)
	// sleep 3 second, tmp length should be 1
	time.Sleep(3 * time.Second)
	if tmp[1] != "" || tmp[0] == "" {
		t.FailNow()
	}
	// sleep 2 second again, tmp length should be 2
	time.Sleep(2 * time.Second)
	if tmp[2] != "" || tmp[1] == "" {
		t.FailNow()
	}
	// status should be "IDLE"
	if manager.JobMap[entryId].Status() != "IDLE" {
		t.FailNow()
	}

	// successTime should be 2, totalTime should be 2
	if manager.JobMap[entryId].SuccessCount != 2 ||
		manager.JobMap[entryId].TotalCount != 2 {
		t.FailNow()
	}
}

//  Test Ignore Panic = True, Manager re-execute panic job even it's panic
func TestIgnorePanic(t *testing.T) {
	resetRunner()
	resetTmp()
	if len(jobReturnHooks) == 0 {
		OnJobReturn(hookAppendResultToTmp)
	}
	manager.SetConfig(CronManagerConfig{true, false, 0, 0})
	fmt.Println("Manager start running: ", time.Now().Minute(), ":", time.Now().Second())
	manager.Start()

	entryId, _ := manager.Add("@every 2s", PanicJob{}, nil)
	// sleep 5 second, tmp length should be 2, Even this is a panic job
	time.Sleep(5 * time.Second)
	if tmp[2] != "" || tmp[1] == "" {
		t.FailNow()
	}

	// status should be "Fail"
	if manager.JobMap[entryId].Status() != "FAIL" {
		t.FailNow()
	}

	// successTime should be 0, totalTime should be 2
	if manager.JobMap[entryId].SuccessCount != 0 ||
		manager.JobMap[entryId].TotalCount != 2 {
		t.FailNow()
	}
}

// Test Ingore Panic = False, Manager won't re-execute panic job
func TestNotIgnorePanic(t *testing.T) {
	resetRunner()
	resetTmp()
	if len(jobReturnHooks) == 0 {
		OnJobReturn(hookAppendResultToTmp)
	}
	manager.SetConfig(CronManagerConfig{false, false, 0, 0})
	fmt.Println("Manager start running: ", time.Now().Minute(), ":", time.Now().Second())
	manager.Start()

	entryId, _ := manager.Add("@every 2s", PanicJob{}, nil)
	// sleep 5 second, tmp length should be 1
	time.Sleep(5 * time.Second)
	if tmp[1] != "" || tmp[0] == "" {
		print("Fail: Two job return")
		t.FailNow()
	}

	// status should be "STOP"
	if manager.JobMap[entryId].Status() != "STOP" {
		print("Fail: Status should be Stop")
		t.FailNow()
	}

	// successTime should be 0, totalTime should be 1
	if manager.JobMap[entryId].SuccessCount != 0 ||
		manager.JobMap[entryId].TotalCount != 1 {
		t.FailNow()
	}
}

// Test Pool size = 2
func TestPoolSize(t *testing.T) {
	resetRunner()
	resetTmp()
	// add job only on time
	if len(jobReturnHooks) == 0 {
		OnJobReturn(hookAppendResultToTmp)
	}
	manager.SetConfig(CronManagerConfig{false, false, 2, 0})
	fmt.Println("Manager start running: ", time.Now().Minute(), ":", time.Now().Second())
	manager.Start()

	entryId, _ := manager.Add("@every 1s", Time3SecJob{}, nil)
	// one running even every 2s
	time.Sleep(5 * time.Second)
	// 0s-1s  idle
	// 1s-4s  first execution
	// 2s-5s  second execution
	// 3s-4s  third block
	// 4s-5s  third execution
	// 5s     forth block
	if tmp[2] != "" || tmp[1] == "" {
		fmt.Print(tmp)
		print("Fail: Two job return")
		t.FailNow()
	}

	// status should be "STOP"
	if manager.JobMap[entryId].Status() != "RUNNING" {
		print("Fail: Status should be running")
		t.FailNow()
	}

	// successTime should be 1, totalTime should be 1
	if manager.JobMap[entryId].SuccessCount != 2 ||
		manager.JobMap[entryId].TotalCount != 2 {
		print("Fail: Status should be running")
		t.FailNow()
	}
}

//// Test Timeout =2
//func TestTimeOut(t *testing.T) {
//	resetRunner()
//	resetTmp()
//	// add job only on time
//	if len(jobReturnHooks) == 0 {
//		OnJobReturn(hookAppendResultToTmp)
//	}
//	manager.SetConfig(CronManagerConfig{true, false, 0, 2})
//	fmt.Println("Manager start running: ", time.Now().Minute(), ":", time.Now().Second())
//	manager.Start()
//
//	entryId, _ := manager.Add("@every 1s", Time3SecJob{}, nil)
//	// one running even every 2s
//	time.Sleep(5 * time.Second)
//	// 0s-1s  idle
//	// 1s-4s  first execution, but timeout in 3s
//	// 2s-5s  second execution, but timeout in 4s
//
//	if tmp[1] != "" || tmp[0] != "" {
//		fmt.Print(tmp)
//		t.FailNow()
//	}
//
//	if manager.JobMap[entryId].SuccessCount != 0 ||
//		manager.JobMap[entryId].TotalCount != 2 {
//		print("Fail: Status should be running")
//		t.FailNow()
//	}
//}

// Test Only One = True, Each job only execute one time , no matter what schedule is
func TestOnlyOne(t *testing.T) {
	resetRunner()
	resetTmp()
	// add job only on time
	if len(jobReturnHooks) == 0 {
		OnJobReturn(hookAppendResultToTmp)
	}
	manager.SetConfig(CronManagerConfig{false, true, 0, 0})
	fmt.Println("Manager start running: ", time.Now().Minute(), ":", time.Now().Second())
	manager.Start()

	entryId, _ := manager.Add("@every 2s", Time3SecJob{}, nil)
	// one running even every 2s
	time.Sleep(6 * time.Second)
	// 0s-2s  idle
	// 2s-5s  first execution
	// 4s-5s  second execution block
	// 5s-6s  second execution, but not finish
	// 6s     third execution block
	if tmp[1] != "" || tmp[0] == "" {
		print("Fail: Two job return")
		t.FailNow()
	}

	// status should be "STOP"
	if manager.JobMap[entryId].Status() != "RUNNING" {
		print("Fail: Status should be running")
		t.FailNow()
	}

	// successTime should be 1, totalTime should be 1
	if manager.JobMap[entryId].SuccessCount != 1 ||
		manager.JobMap[entryId].TotalCount != 1 {
		print("Fail: Status should be running")
		t.FailNow()
	}

	// wait all execution done
	ticker := time.NewTicker(5 * time.Second)
	select {
	case <-ticker.C:
		return
	case <-manager.jobReturnsWithEid:
		return
	}
}

// Test only one = False, parallel job running
func TestNotOnlyOne(t *testing.T) {
	resetRunner()
	resetTmp()
	// add job only on time
	if len(jobReturnHooks) == 0 {
		OnJobReturn(hookAppendResultToTmp)
	}
	manager.SetConfig(CronManagerConfig{false, false, 0, 0})
	fmt.Println("Manager start running: ", time.Now().Minute(), ":", time.Now().Second())
	manager.Start()

	entryId, _ := manager.Add("@every 2s", Time3SecJob{}, nil)
	fmt.Println("Start All job: ", time.Now().Minute(), ":", time.Now().Second())

	// one running even every 2s
	time.Sleep(8 * time.Second)
	// 2s-5s  first execution
	// 4s-7s  second execution
	// 6s-8s  third execution but not finish
	// 8s     fourth execution start

	if tmp[2] != "" || tmp[1] == "" {
		fmt.Println(tmp)
		fmt.Println("Fail: 2 job done, 1 job")
		t.FailNow()
	}

	// status should be "STOP"
	if manager.JobMap[entryId].Status() != "RUNNING" {
		fmt.Println("Fail: Status should be running")
		t.FailNow()
	}

	// successTime should be 1, totalTime should be 1
	if manager.JobMap[entryId].SuccessCount != 2 ||
		manager.JobMap[entryId].TotalCount != 2 {
		fmt.Println("Fail: Status should be running")
		t.FailNow()
	}
}
