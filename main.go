package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Task struct {
	id         int
	createdAt  string
	executedAt string
	errorText  string
}

func main() {
	taskCreator := func(a chan Task, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			createdAt := time.Now().Format(time.RFC3339)
			var executedAt string
			if time.Now().Nanosecond()%2 > 0 {
				executedAt = "Some error occurred"
			} else {
				executedAt = createdAt
			}
			a <- Task{id: int(time.Now().Unix()), createdAt: createdAt, executedAt: executedAt}
		}
	}

	taskWorker := func(t Task) Task {
		createdAt, _ := time.Parse(time.RFC3339, t.createdAt)
		if createdAt.After(time.Now().Add(-20 * time.Second)) {
			t.errorText = "task has been succeeded"
		} else {
			t.errorText = "something went wrong"
		}
		t.executedAt = time.Now().Format(time.RFC3339Nano)
		time.Sleep(time.Millisecond * 150)
		return t
	}

	superChan := make(chan Task, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go taskCreator(superChan, &wg)

	doneTasks := make(chan Task)
	undoneTasks := make(chan Task)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range superChan {
			t = taskWorker(t)
			if t.errorText == "task has been succeeded" {
				doneTasks <- t
			} else {
				undoneTasks <- t
			}
		}
		close(doneTasks)
		close(undoneTasks)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range doneTasks {
			fmt.Println("Done Task ID:", t.id)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range undoneTasks {
			fmt.Printf("Error: Task ID %d, Created At %s, Error: %s\n", t.id, t.createdAt, t.errorText)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	sig := <-c
	fmt.Printf("Received signal: %v\n", sig)
	close(superChan)
	wg.Wait()
}
