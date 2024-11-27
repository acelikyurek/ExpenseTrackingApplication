package main

import (
	"user/internal/handlers"
	"user/internal/services"
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
	router.HandleFunc("/user/login", handlers.HandleLoginRoute(channel)).Methods("POST")
	router.HandleFunc("/user/register", handlers.HandleRegisterRoute(channel)).Methods("POST")

	userService, err := services.NewUserService(connection, mongoURI, databaseName, "users")
	if err != nil {
		log.Fatalf("Failed to initialize user service: %v", err)
	}
	go userService.Start()

	stopChannel := make(chan os.Signal, 1)
	signal.Notify(stopChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Println("Server is starting!")
		log.Fatal(http.ListenAndServe(":8080", router))
	}()

	<-stopChannel
	userService.Stop()
}
