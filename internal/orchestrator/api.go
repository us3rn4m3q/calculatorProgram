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
	ID     int
	Status string
	Result float64
	Expr   string
}

type Task struct {
	ID            int
	Arg1          float64
	Arg2          float64
	Operation     string
	OperationTime int
}

func NewHandler() *Handler {
	return &Handler{
		expressions: make(map[int]Expression),
		tasks:       make(map[int]Task),
		taskChan:    make(chan Task, 100),
		nextID:      1,
	}
}

// Экспортируем CorsMiddleware с большой буквы
func (h *Handler) CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (h *Handler) AddExpression(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid data", http.StatusUnprocessableEntity)
		return
	}

	h.mu.Lock()
	id := h.nextID
	h.nextID++
	h.expressions[id] = Expression{ID: id, Status: "pending", Expr: req.Expression}
	log.Printf("Added expression ID: %d, Expression: %s", id, req.Expression)
	h.mu.Unlock()

	go h.processExpression(id, req.Expression)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	resp := struct {
		Expressions []Expression `json:"expressions"`
	}{}
	for _, expr := range h.expressions {
		resp.Expressions = append(resp.Expressions, expr)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
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
		select {
		case task := <-h.taskChan:
			log.Printf("Sent task to agent: %+v", task)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]Task{"task": task})
		default:
			http.Error(w, "No tasks available", http.StatusNotFound)
		}
	} else if r.Method == http.MethodPost {
		var result struct {
			ID     int     `json:"id"`
			Result float64 `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			log.Printf("Failed to decode task result: %v", err)
			http.Error(w, "Invalid data", http.StatusUnprocessableEntity)
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
		delete(h.tasks, result.ID)

		expr := h.expressions[task.ID]
		expr.Result = result.Result
		expr.Status = "completed"
		h.expressions[task.ID] = expr
		log.Printf("Updated expression ID: %d, Result: %f", task.ID, result.Result)
		h.mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}
}
