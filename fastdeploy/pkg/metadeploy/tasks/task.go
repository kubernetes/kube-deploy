package tasks

import (
	"k8s.io/kube-deploy/fastdeploy/pkg/metadeploy/execution"
	"sort"
)

type Task interface {
	SortKey() (int, string)

	Run(*execution.Context) error
}

const (
	TaskOrderOptions  = 1000
	TaskOrderPackage  = 2000
	TaskOrderFile     = 3000
	TaskOrderTemplate = TaskOrderFile
	TaskOrderSysctl   = 4000
	TaskOrderService  = 5000
)

type TaskList []Task

func (a TaskList) Len() int {
	return len(a)
}
func (a TaskList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a TaskList) Less(i, j int) bool {
	i1, i2 := a[i].SortKey()
	j1, j2 := a[j].SortKey()

	if i1 != j1 {
		return i1 < j1
	}
	return i2 < j2
}

func SortTasks(tasks TaskList) {
	sort.Sort(tasks)
}
