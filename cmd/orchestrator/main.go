package main

import (
	"distributed-calculator/internal/orchestrator"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	os.Setenv("TIME_ADDITION_MS", "2000")
	os.Setenv("TIME_MULTIPLICATIONS_MS", "3000")
	os.Setenv("TIME_SUBTRACTION_MS", "2000")
	os.Setenv("TIME_DIVISIONS_MS", "3000")

	handler := orchestrator.NewHandler()
	mux := http.NewServeMux()

	// Применяем CORS middleware ко всем маршрутам
	mux.HandleFunc("/api/v1/calculate", handler.CorsMiddleware(handler.AddExpression))
	mux.HandleFunc("/api/v1/expressions", handler.CorsMiddleware(handler.GetExpressions))
	mux.HandleFunc("/api/v1/expressions/", handler.CorsMiddleware(handler.GetExpressionByID))
	mux.HandleFunc("/internal/task", handler.CorsMiddleware(handler.HandleTask))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Orchestrator starting on :8080...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
