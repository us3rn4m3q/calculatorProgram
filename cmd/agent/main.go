package main

import (
	"distributed-calculator/internal/agent"
	"log"
	"os"
	"strconv"
)

func main() {
	os.Setenv("COMPUTING_POWER", "3") // 3 горутины по умолчанию
	power, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))

	for i := 0; i < power; i++ {
		go agent.RunWorker()
	}

	log.Println("Agent started with", power, "workers")
	select {} // Бесконечный цикл
}
