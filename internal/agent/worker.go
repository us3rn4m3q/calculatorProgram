package agent

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

// RunWorker starts a worker routine that continuously polls for tasks and processes them
func RunWorker() {
	log.Println("Worker agent started")
	
	// Worker ID for logging
	workerId := time.Now().UnixNano() % 10000
	
	for {
		// Get a task from the orchestrator
		task, err := fetchTask(workerId)
		if err != nil {
			// No task available or error, wait and try again
			time.Sleep(1 * time.Second)
			continue
		}

		// Process the task
		result := compute(workerId, task)
		
		// Send the result back
		err = sendResult(workerId, task.ID, result)
		if err != nil {
			log.Printf("Worker %d: Failed to send result: %v", workerId, err)
			// We could implement retry logic here
			time.Sleep(1 * time.Second)
		}
	}
}

// fetchTask gets a task from the orchestrator
func fetchTask(workerId int64) (*Task, error) {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		log.Printf("Worker %d: Error connecting to orchestrator: %v", workerId, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // No tasks available
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Worker %d: Unexpected status: %d", workerId, resp.StatusCode)
		return nil, err
	}

	var data struct {
		Task Task `json:"task"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("Worker %d: Error decoding task: %v", workerId, err)
		return nil, err
	}

	log.Printf("Worker %d: Received task ID: %d (%s)", workerId, data.Task.ID, data.Task.Operation)
	return &data.Task, nil
}

// compute performs the mathematical operation
func compute(workerId int64, task *Task) float64 {
	log.Printf("Worker %d: Computing %f %s %f (wait time: %dms)", 
		workerId, task.Arg1, task.Operation, task.Arg2, task.OperationTime)
	
	// Simulate processing time
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	
	// Perform the operation
	var result float64
	switch task.Operation {
	case "+":
		result = task.Arg1 + task.Arg2
	case "-":
		result = task.Arg1 - task.Arg2
	case "*":
		result = task.Arg1 * task.Arg2
	case "/":
		// Handle division by zero
		if task.Arg2 == 0 {
			log.Printf("Worker %d: Division by zero error!", workerId)
			return 0
		}
		result = task.Arg1 / task.Arg2
	default:
		log.Printf("Worker %d: Unknown operation: %s", workerId, task.Operation)
		return 0
	}
	
	log.Printf("Worker %d: Result: %f", workerId, result)
	return result
}

// sendResult sends the computation result back to the orchestrator
func sendResult(workerId int64, id int, result float64) error {
	// Prepare the result data
	reqBody, err := json.Marshal(map[string]interface{}{
		"id":     id,
		"result": result,
	})
	
	if err != nil {
		return err
	}
	
	// Send the result
	resp, err := http.Post(
		"http://localhost:8080/internal/task", 
		"application/json", 
		bytes.NewBuffer(reqBody),
	)
	
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check for success
	if resp.StatusCode != http.StatusOK {
		log.Printf("Worker %d: Got unexpected status: %d", workerId, resp.StatusCode)
		return err
	}
	
	log.Printf("Worker %d: Result for task %d sent successfully", workerId, id)
	return nil
}