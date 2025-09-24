package core

import (
	"fmt"
	"testing"
	"time"

	"paleotronic.com/log"
)

type ThingWhatDoes struct {
	TaskPerformer
	running bool
}

func (th *ThingWhatDoes) handleTask(task *Task) (interface{}, error) {
	log.Printf("Got a task: action = %s", task.Action)
	switch task.Action {
	case "stop":
		th.running = false
		return "ok", nil
	}
	return "", fmt.Errorf("unrecognised task: %s", task.Action)
}

func (th *ThingWhatDoes) Run(t *testing.T) {

	go func() {
		th.running = true
		for th.running {
			t.Log("running")
			time.Sleep(1 * time.Second)
			th.HandleTasks(th.handleTask)
		}
	}()

}

func TestTaskExecution(t *testing.T) {

	thing := &ThingWhatDoes{}
	go thing.Run(t)

	tt := NewTask("stop")
	resp := tt.Request(thing)

	if resp.Err != nil {
		t.Fatal("Expected resp.Err to be nil")
	}

	if thing.running {
		t.Fatal("Expected thing to stop running")
	}

	t.Fail()

}
