package restcron

import "sort"

type CronHook struct {
	order int
	f func(runReturn *JobRunReturn)
}

func OnJobReturn(f func(runReturn *JobRunReturn), order ...int){
	jobReturnHooks = jobReturnHooks.Add(f, order...)
}

type CronHooks []CronHook


var jobReturnHooks CronHooks


func(h CronHooks) Run(runReturn *JobRunReturn) {
	sort.Sort(h)
	for _, hook := range h {
		hook.f(runReturn)
	}
}


func(h CronHooks) Add(fn func(runReturn *JobRunReturn), order ...int) CronHooks {
	o := 1
	if len(order) > 0 {
		o = order[0]
	}
	return append(h, CronHook{o, fn})
}

// Sorting function
func (h CronHooks) Len() int {
	return len(h)
}

// Sorting function
func (h CronHooks) Less(i, j int) bool {
	return h[i].order < h[j].order
}

// Sorting function
func (h CronHooks) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}


