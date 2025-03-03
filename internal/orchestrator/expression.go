package orchestrator

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// processExpression handles parsing and processing of mathematical expressions
func (h *Handler) processExpression(id int, expr string) {
	tokens := strings.Fields(expr)
	if len(tokens) < 3 || len(tokens)%2 == 0 {
		h.mu.Lock()
		h.expressions[id] = Expression{
			ID:     id,
			Status: "error",
			Result: 0,
			Expr:   expr,
		}
		h.mu.Unlock()
		log.Printf("Invalid expression format: %s", expr)
		return
	}

	// Parse the expression
	result, err := h.evaluateExpression(id, tokens)
	if err != nil {
		h.mu.Lock()
		h.expressions[id] = Expression{
			ID:     id,
			Status: "error",
			Result: 0,
			Expr:   expr,
		}
		h.mu.Unlock()
		log.Printf("Error evaluating expression: %v", err)
	}

	log.Printf("Expression %d (%s) evaluated to %f", id, expr, result)
}

// evaluateExpression processes a simple math expression according to operator precedence
func (h *Handler) evaluateExpression(id int, tokens []string) (float64, error) {
	// For simplicity, we'll implement a basic expression evaluator
	// This is a simplified version - a more robust parser would be needed for complex expressions
	
	// First pass: handle multiplication and division
	for i := 1; i < len(tokens); i += 2 {
		if tokens[i] == "*" || tokens[i] == "/" {
			arg1, err1 := strconv.ParseFloat(tokens[i-1], 64)
			arg2, err2 := strconv.ParseFloat(tokens[i+1], 64)
			
			if err1 != nil || err2 != nil {
				return 0, err1
			}
			
			opTime := getOperationTime(tokens[i])
			task := Task{
				ID:            id, 
				Arg1:          arg1, 
				Arg2:          arg2, 
				Operation:     tokens[i], 
				OperationTime: opTime,
			}
			
			h.mu.Lock()
			h.tasks[id] = task
			h.taskChan <- task
			h.mu.Unlock()
			
			// For now, just handle one operation at a time
			// A more complex evaluator would handle full expressions
			return 0, nil
		}
	}

	// If no multiplication/division, handle addition/subtraction
	arg1, err1 := strconv.ParseFloat(tokens[0], 64)
	arg2, err2 := strconv.ParseFloat(tokens[2], 64)
	
	if err1 != nil || err2 != nil {
		return 0, err1
	}
	
	opTime := getOperationTime(tokens[1])
	task := Task{
		ID:            id, 
		Arg1:          arg1, 
		Arg2:          arg2, 
		Operation:     tokens[1], 
		OperationTime: opTime,
	}
	
	h.mu.Lock()
	h.tasks[id] = task
	h.taskChan <- task
	h.mu.Unlock()
	
	return 0, nil
}

// getOperationTime returns the configured processing time for each operation type
func getOperationTime(op string) int {
	switch op {
	case "+":
		return atoi(os.Getenv("TIME_ADDITION_MS"), 2000)
	case "-":
		return atoi(os.Getenv("TIME_SUBTRACTION_MS"), 2000)
	case "*":
		return atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"), 3000)
	case "/":
		return atoi(os.Getenv("TIME_DIVISIONS_MS"), 3000)
	default:
		return 1000
	}
}

// atoi safely converts string to int with a default value
func atoi(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return i
}