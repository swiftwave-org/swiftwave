package main

import "github.com/swiftwave-org/swiftwave/task_queue"

type Data struct {
	x int `json:"x"`
	y int `json:"y"`
}

func testAddition(data Data) error {
	return nil
}

func testSubtraction(data Data) error {
	return nil
}

func TestTaskQueue() {
	client, err := task_queue.NewClient(task_queue.Options{Type: task_queue.Local})
	if err != nil {
		panic(err)
	}
	err = client.RegisterFunction("addition", testAddition)
	if err != nil {
		panic(err)
	}
	err = client.RegisterFunction("subtraction", testSubtraction)
	if err != nil {
		panic(err)
	}
}
