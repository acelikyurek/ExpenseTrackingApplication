package services

import (
	"user/internal/middlewares"
	"user/internal/models"
	"user/internal/repositories"
	"log"
	"os"
	"encoding/json"
	"context"

	"github.com/google/uuid"	
	"github.com/streadway/amqp"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	connection	*amqp.Connection
	channel		*amqp.Channel
	mongoDBRepo	*repositories.MongoDBRepository
}

func NewUserService(connection *amqp.Connection, uri string, databaseName string, collectionName string) (*UserService, error) {
	channel, err := connection.Channel()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	repo, err := repositories.NewMongoDBRepository(uri, databaseName, collectionName)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UserService{
		connection: 	connection,
		channel: 		channel,
		mongoDBRepo: 	repo,
	}, nil
}

func (s *UserService) Start() {
	defer s.connection.Close()
	defer s.channel.Close()
	
	rabbitmqURI := os.Getenv("RABBITMQ_URI")
	connection, err := amqp.Dial(rabbitmqURI)
	if err != nil {
		log.Println(err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Println(err)
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare("userQueue", false, false, false, false, nil)
	if err != nil {
		log.Println(err)
	}

	messages, err := s.channel.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println(err)
	}

	for message := range messages {
		action, ok := message.Headers["action"].(string)
		if !ok {
			continue
		}
		switch action {
			case "Login":
				s.HandleLogin(message.Body, message.ReplyTo, message.CorrelationId)
			case "Register":
				s.HandleRegister(message.Body, message.ReplyTo, message.CorrelationId)
			default:
				log.Printf("Action (%s) is unknown!", action)
		}
	}
}

func (s *UserService) Stop() {
	log.Println("User service is being stopping.")
	if err := s.channel.Close(); err != nil {
		log.Println(err)
	}
	log.Println("User service is stopped.")
}

type loginServiceRequest struct {
	Action   string `json:"action"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginServiceResponse struct {
	Message string 	`json:"message"`
	Success bool 	`json:"success"`
	Token  	string 	`json:"token"`
}

func (s *UserService) HandleLogin(data []byte, replyTo string, correlationId string) {
	loginServiceRequestData := loginServiceRequest{}
	err := json.Unmarshal(data, &loginServiceRequestData)
    if err != nil {
		log.Println(err)
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"LoginResponse", 
			loginServiceResponse{
				Message: "An error occured!",
				Success: false,
				Token: "",
			},
		)
		return
    }
	
	filter := map[string]interface{}{"email": loginServiceRequestData.Email}
	result, err := s.mongoDBRepo.Find(context.Background(), filter)
	if !result.Success {
		log.Println(err)
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"LoginResponse", 
			loginServiceResponse{
				Message: "An error occured!",
				Success: false,
				Token: "",
			},
		)
		return
	} 
	if len(result.Data) == 0 {
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"LoginResponse", 
			loginServiceResponse{
				Message: "User not found!",
				Success: false,
				Token: "",
			},
		)
		return
	}

	var user models.User
	jsonData, err := json.Marshal(result.Data[0])
    if err != nil {
		log.Println(err)
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"LoginResponse", 
			loginServiceResponse{
				Message: "An error occured!",
				Success: false,
				Token: "",
			},
		)
		return
    }
    json.Unmarshal(jsonData, &user)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginServiceRequestData.Password)); err != nil {
		log.Println(err)
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"LoginResponse", 
			loginServiceResponse{
				Message: "Password is invalid!",
				Success: false,
				Token: "",
			},
		)
		return
	}

	token, err := middlewares.CreateToken(user)
	if err != nil {
		log.Println(err)
        SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"LoginResponse", 
			loginServiceResponse{
				Message: "An error occured!",
				Success: false,
				Token: "",
			},
		)
        return
    }

	SendResponse(
		s.channel, 
		replyTo, 
		correlationId, 
		"LoginResponse", 
		loginServiceResponse{
			Message: "Login is successful!",
			Success: true,
			Token: token,
	    },
	)
}

type registerServiceRequest struct {
	Action   string `json:"action"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name 	 string `json:"name"`
}

type registerServiceResponse struct {
	Message string 	`json:"message"`
	Success bool 	`json:"success"`
}

func (s *UserService) HandleRegister(data []byte, replyTo string, correlationId string) {	
	registerServiceRequestData := registerServiceRequest{}
	err := json.Unmarshal(data, &registerServiceRequestData)
    if err != nil {
        log.Println(err)
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"RegisterResponse", 
			registerServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
        return
    }
	
	filter := map[string]interface{}{"email": registerServiceRequestData.Email}
	findResult, err := s.mongoDBRepo.Find(context.Background(), filter)
	if !findResult.Success {
		log.Println(err)
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"RegisterResponse", 
			registerServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
		return
	}

	if len(findResult.Data) != 0 {
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"RegisterResponse", 
			registerServiceResponse{
				Message: "Email is already in use!",
				Success: false,
			},
		)
		return
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(registerServiceRequestData.Password), bcrypt.DefaultCost)
    if err != nil {
		log.Println(err)
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"RegisterResponse", 
			registerServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
		return
    }

	userId := uuid.New().String()

	newUser := models.User{
		UserId: userId,
		Email: registerServiceRequestData.Email,
		Password: string(hashedPasswordBytes),
		Name: registerServiceRequestData.Name,
	}
	insertionResult, err := s.mongoDBRepo.Insert(context.Background(), newUser)
	if !insertionResult.Success {
		log.Println(err)
		SendResponse(
			s.channel, 
			replyTo, 
			correlationId, 
			"RegisterResponse", 
			registerServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
		return
	}

	SendResponse(
		s.channel, 
		replyTo, 
		correlationId, 
		"RegisterResponse", 
		registerServiceResponse{
			Message: "Registration is successful!",
			Success: true,
		},
	)
}
