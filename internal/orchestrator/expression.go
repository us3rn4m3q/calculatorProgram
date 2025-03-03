package orchestrator

import (
	"os"
	"strconv"
	"strings"
)

func (h *Handler) processExpression(id int, expr string) {
	tokens := strings.Fields(expr)
	if len(tokens) < 3 {
		h.mu.Lock()
		h.expressions[id] = Expression{ID: id, Status: "error", Expr: expr}
		h.mu.Unlock()
		return
	}

	// Пример: "2 + 3" или "2 + 2 * 2"
	// Сначала выполняем умножение/деление
	for i := 1; i < len(tokens)-1; i += 2 {
		if tokens[i] == "*" || tokens[i] == "/" {
			arg1, _ := strconv.ParseFloat(tokens[i-1], 64)
			arg2, _ := strconv.ParseFloat(tokens[i+1], 64)
			opTime := getOperationTime(tokens[i])
			task := Task{ID: id, Arg1: arg1, Arg2: arg2, Operation: tokens[i], OperationTime: opTime}
			h.mu.Lock()
			h.tasks[id] = task
			h.taskChan <- task
			h.mu.Unlock()
			return // Пока обрабатываем только одну операцию за раз
		}
	}

	// Сложение/вычитание
	arg1, _ := strconv.ParseFloat(tokens[0], 64)
	arg2, _ := strconv.ParseFloat(tokens[2], 64)
	opTime := getOperationTime(tokens[1])
	task := Task{ID: id, Arg1: arg1, Arg2: arg2, Operation: tokens[1], OperationTime: opTime}
	h.mu.Lock()
	h.tasks[id] = task
	h.taskChan <- task
	h.mu.Unlock()
}

func getOperationTime(op string) int {
	switch op {
	case "+":
		return atoi(os.Getenv("TIME_ADDITION_MS"))
	case "-":
		return atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	case "*":
		return atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	case "/":
		return atoi(os.Getenv("TIME_DIVISIONS_MS"))
	default:
		return 1000
	}
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
