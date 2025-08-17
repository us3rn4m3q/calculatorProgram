# Project Description

This project is a web service that accepts arithmetic expressions via HTTP requests and returns the result of their evaluation.  
To use it, you need to send a POST request with the expression in JSON format.

---

## Getting Started

### Run the service

To start the project, use the following command in the console:

bash
go run ./cmd/calc_service/...
If you are using a JetBrains IDE, you can also start the service using the built-in run button.

Request Examples
PowerShell
To send a request via PowerShell:

powershell
Copy
Edit
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/calculate" `
                  -Method POST `
                  -ContentType "application/json" `
                  -Body '{"expression": "2-2"}'
Command Line
To send a request via standard command line:

bash
Copy
Edit
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
    "expression": "2/0"
}'
Testing
To run tests, execute the following command:

bash
Copy
Edit
go test ./...
