package machine

import (
	"time"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"github.com/kubernetes-incubator/apiserver-builder/pkg/controller"
)

func (c *MachineController) RunAsync(stopCh <-chan struct{}) {
	for _, q := range c.Informers.WorkerQueues {
		c.StartWorkerQueue(q,stopCh)
	}
	controller.GetDefaults(c.controller).Run(stopCh)
}

// StartWorkerQueue schedules a routine to continuously process Queue messages
// until shutdown is closed
func (c *MachineController) StartWorkerQueue(q *controller.QueueWorker, shutdown <-chan struct{}) {
	defer runtime.HandleCrash()
	glog.Infof("Start %s Queue", q.Name)

	// Every second, process all messages in the Queue until it is time to shutdown
	go wait.Until(func(){c.ProcessAllMessages(q)}, time.Second, shutdown)

	go func() {
		<-shutdown

		// Stop accepting messages into the Queue
		glog.V(1).Infof("Shutting down %s Queue\n", q.Name)
		q.Queue.ShutDown()
	}()
}

// ProcessAllMessages tries to process all messages in the Queue
func (c *MachineController) ProcessAllMessages(q *controller.QueueWorker) {
	for done := false; !done; {
		// Process all messages in the Queue
		done = c.ProcessMessage(q)
	}
}

// ProcessMessage tries to process the next message in the Queue, and requeues on an error
func (c *MachineController) ProcessMessage(q *controller.QueueWorker) bool {
	key, quit := q.Queue.Get()
	if quit {
		// Queue is empty
		return true
	}

	// Create sync channel if needed.
	syncChan, ok := c.controller.syncChans[key.(string)]
	if !ok {
		c.controller.syncChans[key.(string)] = make(chan string, 1)
		syncChan = c.controller.syncChans[key.(string)]
	}

	go func() {
		defer q.Queue.Done(key)
		// sync against same machine resource. Allow only one reconcilation action in progress per
		// machine resource.
		syncChan <- key.(string)
		defer func(){
			<-syncChan
			}()

		// Do the work
		err := q.Reconcile(key.(string))
		if err == nil {
			// Success.  Clear the requeue count for this key.
			q.Queue.Forget(key)
			return
		}

		// Error.  Maybe retry if haven't hit the limit.
		if q.Queue.NumRequeues(key) < q.MaxRetries {
			glog.V(4).Infof("Error handling %s Queue message %v: %v", q.Name, key, err)
			q.Queue.AddRateLimited(key)
			return
		}

		glog.V(4).Infof("Too many retries for %s Queue message %v: %v", q.Name, key, err)
		q.Queue.Forget(key)

	}()

	return false
}