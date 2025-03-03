package orchestrator

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Handler struct {
	expressions map[int]Expression
	tasks       map[int]Task
	taskChan    chan Task
	mu          sync.Mutex
	nextID      int
}

type Expression struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
	Expr   string  `json:"expr"`
}

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

func NewHandler() *Handler {
	return &Handler{
		expressions: make(map[int]Expression),
		tasks:       make(map[int]Task),
		taskChan:    make(chan Task, 100),
		nextID:      1,
	}
}

func (h *Handler) CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (h *Handler) AddExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Expression) == "" {
		http.Error(w, "Expression cannot be empty", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	id := h.nextID
	h.nextID++
	h.expressions[id] = Expression{
		ID:     id,
		Status: "pending",
		Result: 0,
		Expr:   req.Expression,
	}
	log.Printf("Added expression ID: %d, Expression: %s", id, req.Expression)
	h.mu.Unlock()

	go h.processExpression(id, req.Expression)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	expressions := make([]Expression, 0, len(h.expressions))
	for _, expr := range h.expressions {
		expressions = append(expressions, expr)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]Expression{"expressions": expressions})
}

func (h *Handler) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	expr, exists := h.expressions[id]
	h.mu.Unlock()

	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]Expression{"expression": expr})
}

func (h *Handler) HandleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Worker agent requests a task
		select {
		case task := <-h.taskChan:
			log.Printf("Sent task to agent: %+v", task)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]Task{"task": task})
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "No tasks available"})
		}
	} else if r.Method == http.MethodPost {
		// Worker agent sends back a result
		var result struct {
			ID     int     `json:"id"`
			Result float64 `json:"result"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			log.Printf("Failed to decode task result: %v", err)
			http.Error(w, "Invalid data format", http.StatusBadRequest)
			return
		}

		h.mu.Lock()
		task, exists := h.tasks[result.ID]
		if !exists {
			h.mu.Unlock()
			log.Printf("Task not found for ID: %d", result.ID)
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		
		// Remove the task from the pending tasks
		delete(h.tasks, result.ID)

		// Update the expression with the result
		expr, exprExists := h.expressions[task.ID]
		if exprExists {
			expr.Result = result.Result
			expr.Status = "completed"
			h.expressions[task.ID] = expr
			log.Printf("Updated expression ID: %d, Result: %f", task.ID, result.Result)
		}
		h.mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}