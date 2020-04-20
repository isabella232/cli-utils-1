// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package taskrunner

import (
	"sync"
	"testing"
	"time"

	"sigs.k8s.io/cli-utils/pkg/apply/event"
	"sigs.k8s.io/cli-utils/pkg/object"
)

func TestWaitTask_TimeoutTriggered(t *testing.T) {
	task := NewWaitTask([]object.ObjMetadata{}, AllCurrent, 2*time.Second)

	eventChannel := make(chan event.Event)
	taskContext := NewTaskContext(eventChannel)
	defer close(eventChannel)

	task.Start(taskContext)

	timer := time.NewTimer(3 * time.Second)

	select {
	case res := <-taskContext.TaskChannel():
		if res.Err == nil || !IsTimeoutError(res.Err) {
			t.Errorf("expected timeout error, but got %v", res.Err)
		}
		return
	case <-timer.C:
		t.Errorf("expected timeout to trigger, but it didn't")
	}
}

func TestWaitTask_TimeoutCancelled(t *testing.T) {
	task := NewWaitTask([]object.ObjMetadata{}, AllCurrent, 2*time.Second)

	eventChannel := make(chan event.Event)
	taskContext := NewTaskContext(eventChannel)
	defer close(eventChannel)

	task.Start(taskContext)
	task.ClearTimeout()
	timer := time.NewTimer(3 * time.Second)

	select {
	case res := <-taskContext.TaskChannel():
		t.Errorf("didn't expect timeout error, but got %v", res.Err)
	case <-timer.C:
		return
	}
}

func TestWaitTask_SingleTaskResult(t *testing.T) {
	task := NewWaitTask([]object.ObjMetadata{}, AllCurrent, 2*time.Second)

	eventChannel := make(chan event.Event)
	taskContext := NewTaskContext(eventChannel)
	taskContext.taskChannel = make(chan TaskResult, 10)
	defer close(eventChannel)

	var completeWg sync.WaitGroup

	for i := 0; i < 10; i++ {
		completeWg.Add(1)
		go func() {
			defer completeWg.Done()
			task.complete(taskContext)
		}()
	}
	completeWg.Wait()

	<-taskContext.TaskChannel()

	timer := time.NewTimer(4 * time.Second)

	select {
	case <-taskContext.TaskChannel():
		t.Errorf("expected only one result on taskChannel, but got more")
	case <-timer.C:
		return
	}
}