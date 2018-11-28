package croner

import (
	"testing"
	"time"
	"fmt"
)

var tmp = [5]string{}
var manager = NewCronManager(CronManagerConfig{true, false, 0, 0})

func resetTmp() {
	tmp = [5]string{}
}

func resetRunner() {
	if manager.running{
		manager.RemoveAll()
		manager.Stop()
	}
}

type PanicJob struct {}

func (j PanicJob) Run() JobRunReturn{
	panic("hello, i am a panic job")
}

type GoodJob struct {}

func (j GoodJob) Run() JobRunReturn{
	return JobRunReturn{"Hello , I am a good job", nil}
}

type TimeOutJob struct {}

func (j TimeOutJob) Run() JobRunReturn {
	time.Sleep(5 * time.Second)
	return JobRunReturn{"Hello , I am a timeout job", nil }
}


func hookAppendResultToTmp(runReturn *JobRunReturnWithEid) {
	for i, v := range tmp{
		if v == ""{
			tmpStr := fmt.Sprintf("%v", runReturn.Value)
			tmp[i] = tmpStr
			break
		}
	}
}

// test simple logic
func TestRunning(t *testing.T) {
	resetRunner()
	resetTmp()
	OnJobReturn(hookAppendResultToTmp)
	manager.Start()

	entryId, _ := manager.Add("@every 2s", GoodJob{}, nil)
	// sleep 3 second, tmp length should be 1
	time.Sleep(3 * time.Second)
	if tmp[1] != "" || tmp[0] == ""{
		t.FailNow()
	}
	// sleep 2 second again, tmp length should be 2
	time.Sleep(2 * time.Second)
	if tmp[2] != "" || tmp[1] == ""{
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

//  Test Ignore Panic = True
func TestIgnorePanic(t *testing.T) {
	resetRunner()
	resetTmp()
	OnJobReturn(hookAppendResultToTmp)
	manager.SetConfig(CronManagerConfig{true, false, 0, 0})
	manager.Start()

	entryId, _ := manager.Add("@every 2s", PanicJob{}, nil)
	// sleep 5 second, tmp length should be 2, Even this is a panic job
	time.Sleep(5 * time.Second)
	if tmp[2] != "" || tmp[1] == ""{
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

// Test Ingore Panic = False
func TestNotIgnorePanic(t *testing.T) {
	resetRunner()
	resetTmp()
	OnJobReturn(hookAppendResultToTmp)
	manager.SetConfig(CronManagerConfig{false, false, 0, 0})
	manager.Start()

	entryId, _ := manager.Add("@every 2s", PanicJob{}, nil)
	// sleep 5 second, tmp length should be 1
	time.Sleep(5 * time.Second)
	if tmp[1] != "" || tmp[0] == ""{
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


// Test Only One = True
func TestOnlyOne(t *testing.T) {
	resetRunner()
	resetTmp()
	OnJobReturn(hookAppendResultToTmp)
	manager.SetConfig(CronManagerConfig{false, true, 0, 0})
	manager.Start()

	entryId, _ :=manager.Add("@every 2s", TimeOutJob{}, nil)
	// only one running even every 2s
	time.Sleep(8 * time.Second)
	if tmp[1] != "" || tmp[0] == ""{
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
}






