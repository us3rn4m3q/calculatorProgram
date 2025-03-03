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

func RunWorker() {
	for {
		resp, err := http.Get("http://localhost:8080/internal/task")
		if err != nil || resp.StatusCode == http.StatusNotFound {
			time.Sleep(1 * time.Second)
			continue
		}

		var data struct {
			Task Task `json:"task"`
		}
		json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()

		task := data.Task
		log.Printf("Agent received task: %+v\n", task) // Отладка
		result := compute(task)
		log.Printf("Agent computed result: %f\n", result) // Отладка
		sendResult(task.ID, result)
	}
}

func compute(task Task) float64 {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}

func sendResult(id int, result float64) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"id":     id,
		"result": result,
	})
	resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println("Failed to send result:", err)
		return
	}
	resp.Body.Close()
}
