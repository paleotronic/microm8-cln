package core

// Task is a parametized request a VM must execute
// it must return a value to the response channel once done
type TaskResponse struct {
	Err   error
	Value interface{}
}

type Task struct {
	Action    string // scoped action
	Arguments []interface{}
	Response  chan *TaskResponse
}

// Creates a new Task with an embedded response channel
func NewTask(action string, args ...interface{}) *Task {
	return &Task{
		Action:    action,
		Arguments: args,
		Response:  make(chan *TaskResponse, 16),
	}
}

// Execute - perform the task and await the response
func (t *Task) Request(target Taskable) *TaskResponse {
	target.TaskRequest(t)
	//tt := time.After(1 * time.Second)
	//select {
	resp := <-t.Response
	close(t.Response)
	return resp
	//case <-tt:
	//	return &TaskResponse{
	//		Value: nil,
	//		Err:   errors.New(fmt.Sprintf("timeout waiting for task %s [%+v]", t.Action, t.Arguments)),
	//	}
	//}
}

// TaskPerformer is an interface for something that executes tasks
type Taskable interface {
	TaskRequest(t *Task)
}

type TaskPerformer struct {
	IncomingTasks chan *Task
}

func (tp *TaskPerformer) ExecuteRequest(action string, args ...interface{}) (interface{}, error) {
	t := NewTask(action, args...)
	r := t.Request(tp)
	return r.Value, r.Err
}

func (tp *TaskPerformer) ExecuteRequestImm(action string, args ...interface{}) (interface{}, error) {
	t := NewTask(action, args...)
	r := t.Request(tp)
	return r.Value, r.Err
}

func (tp *TaskPerformer) TaskRequest(t *Task) {
	if tp.IncomingTasks == nil {
		tp.IncomingTasks = make(chan *Task, 128)
	}
	tp.IncomingTasks <- t
}

type TaskHandler func(t *Task) (interface{}, error)

func (tp *TaskPerformer) HandleTasks(h TaskHandler) {

	if tp.IncomingTasks == nil {
		tp.IncomingTasks = make(chan *Task, 128)
	}

	for len(tp.IncomingTasks) > 0 {
		task := <-tp.IncomingTasks
		value, err := h(task)
		task.Response <- &TaskResponse{
			Err:   err,
			Value: value,
		}
	}

}
