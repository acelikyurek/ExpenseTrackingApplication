package main

import (
	"expense/internal/middlewares"
	"expense/internal/handlers"
	"expense/internal/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	databaseName := os.Getenv("DATABASE_NAME")
	rabbitmqURI := os.Getenv("RABBITMQ_URI")

	connection, err := amqp.Dial(rabbitmqURI)
	if err != nil {
		log.Fatalf("RabbitMQ connection is failed: %v", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Channel creation is failed: %v", err)
	}
	defer channel.Close()

	router := mux.NewRouter()
	router.HandleFunc("/expense", middlewares.AuthMiddleware(handlers.HandleAddExpenseRoute(channel))).Methods("POST")
	router.HandleFunc("/expense", middlewares.AuthMiddleware(handlers.HandleGetExpenseRoute(channel))).Methods("GET")
	router.HandleFunc("/expense", middlewares.AuthMiddleware(handlers.HandleUpdateExpenseRoute(channel))).Methods("PUT")
	router.HandleFunc("/expense", middlewares.AuthMiddleware(handlers.HandleRemoveExpenseRoute(channel))).Methods("DELETE")

	expenseService, err := services.NewExpenseService(connection, mongoURI, databaseName, "expenses")
	if err != nil {
		log.Fatalf("Failed to initialize expense service: %v", err)
	}
	go expenseService.Start()

	stopChannel := make(chan os.Signal, 1)
	signal.Notify(stopChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Println("Server is starting!")
		log.Fatal(http.ListenAndServe(":8080", router))
	}()

	<-stopChannel
	expenseService.Stop()
}
