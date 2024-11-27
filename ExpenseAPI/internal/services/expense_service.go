package services

import (
	"expense/internal/models"
	"expense/internal/repositories"
	"log"
	"os"
	"encoding/json"
	"context"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type ExpenseService struct {
	connection	*amqp.Connection
	channel		*amqp.Channel
	mongoDBRepo	*repositories.MongoDBRepository
}

func NewExpenseService(connection *amqp.Connection, uri string, databaseName string, collectionName string) (*ExpenseService, error) {
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

	return &ExpenseService{
		connection: 	connection,
		channel: 		channel,
		mongoDBRepo: 	repo,
	}, nil
}

func (s *ExpenseService) Start() {
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

	queue, err := channel.QueueDeclare("expenseQueue", false, false, false, false, nil)
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
			case "GetExpense":
				s.HandleGetExpense(message.Body, message.ReplyTo, message.CorrelationId)
			case "AddExpense":
				s.HandleAddExpense(message.Body, message.ReplyTo, message.CorrelationId)
			case "UpdateExpense":
				s.HandleUpdateExpense(message.Body, message.ReplyTo, message.CorrelationId)
			case "RemoveExpense":
				s.HandleRemoveExpense(message.Body, message.ReplyTo, message.CorrelationId)
			default:
				log.Printf("Action (%s) is unknown!", action)
		}
	}
}

func (s *ExpenseService) Stop() {
	log.Println("Expense service is being stopping.")
	if err := s.channel.Close(); err != nil {
		log.Println(err)
	}
	log.Println("Expense service is stopped.")
}

type getExpenseServiceRequest struct {
	Action   string 				`json:"action"`
	UserId	 string 				`json:"userId"`
	Filter   map[string]interface{} `json:"filter"`
}

type getExpenseServiceResponse struct {
	Message  string 		  			`json:"message"`
	Success  bool 			  			`json:"success"`
	Expenses []map[string]interface{}	`json:"expenses"`
}

func (s *ExpenseService) HandleGetExpense(data []byte, replyTo string, correlationId string) {
	getExpenseServiceRequestData := getExpenseServiceRequest{}
	err := json.Unmarshal(data, &getExpenseServiceRequestData)
    if err != nil {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"GetExpenseResponse",
			getExpenseServiceResponse{
				Message: "An error occured!",
				Success: false,
				Expenses: nil,
			},
		)
        return
    }
	
	filter := map[string]interface{}{
		"userId": getExpenseServiceRequestData.UserId,
	}
	for key, value := range getExpenseServiceRequestData.Filter {
		filter[key] = value
	}
	result, err := s.mongoDBRepo.Find(context.Background(), filter)
	if !result.Success {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"GetExpenseResponse",
			getExpenseServiceResponse{
				Message: "An error occured!",
				Success: false,
				Expenses: nil,
			},
		)
        return
	}
	if len(result.Data) == 0 {
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"GetExpenseResponse",
			getExpenseServiceResponse{
				Message: "No expenses found!",
				Success: true,
				Expenses: []map[string]interface{}{},
			},
		)
		return
	}

	expenses := result.Data
	SendResponse(
		s.channel, 
		replyTo, 
		correlationId, 
		"GetExpenseResponse", 
		getExpenseServiceResponse{
			Message: "Operation is successful!",
			Success: true,
			Expenses: expenses,
		},
	)
}

type addExpenseServiceRequest struct {
	UserId		string 	`json:"userId"`
	Action   	string 	`json:"action"`
	Description string 	`json:"description"`
	Amount 		float32 `json:"amount"`
	Category 	string 	`json:"category"`
}

type addExpenseServiceResponse struct {
	Message string 		   `json:"message"`
	Success bool 		   `json:"success"`
}

func (s *ExpenseService) HandleAddExpense(data []byte, replyTo string, correlationId string) {
	addExpenseServiceRequestData := addExpenseServiceRequest{}
	err := json.Unmarshal(data, &addExpenseServiceRequestData)
    if err != nil {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"AddExpenseResponse",
			addExpenseServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
        return
    }
	
	expenseId := uuid.New().String()

	expense := models.Expense{
		ExpenseId:   expenseId,
		UserId:      addExpenseServiceRequestData.UserId,
		Description: addExpenseServiceRequestData.Description,
		Amount:      addExpenseServiceRequestData.Amount,
		Category:    addExpenseServiceRequestData.Category,
	}

	result, err := s.mongoDBRepo.Insert(context.Background(), expense)
	if !result.Success {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"AddExpenseResponse",
			addExpenseServiceResponse{
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
		"AddExpenseResponse",
		addExpenseServiceResponse{
			Message: "Operation is successful!",
			Success: true,
		},
	)
}

type updateExpenseServiceRequest struct {
	Action   	string 	`json:"action"`
	UserId		string 	`json:"userId"`
	ExpenseId	string 	`json:"expenseId"`
	Description string 	`json:"description"`
	Amount 		float32 `json:"amount"`
	Category 	string 	`json:"category"`
}

type updateExpenseServiceResponse struct {
	Message string 		   `json:"message"`
	Success bool 		   `json:"success"`
}

func (s *ExpenseService) HandleUpdateExpense(data []byte, replyTo string, correlationId string) {
	updateExpenseServiceRequestData := updateExpenseServiceRequest{}
	err := json.Unmarshal(data, &updateExpenseServiceRequestData)
    if err != nil {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"UpdateExpenseResponse",
			updateExpenseServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
        return
    }
	
	filter := map[string]interface{}{
		"userId": updateExpenseServiceRequestData.UserId,
		"expenseId": updateExpenseServiceRequestData.ExpenseId,
	}
	result, err := s.mongoDBRepo.Find(context.Background(), filter)
	if !result.Success {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"UpdateExpenseResponse",
			updateExpenseServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
        return
	}
	if len(result.Data) == 0 {
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"UpdateExpenseResponse",
			updateExpenseServiceResponse{
				Message: "Expense not found!",
				Success: false,
			},
		)
		return
	}

	expense := models.Expense{}
	jsonData, err := json.Marshal(result.Data[0])
    if err != nil {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"UpdateExpenseResponse",
			updateExpenseServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
        return
    }
	json.Unmarshal(jsonData, &expense)
	expense.Description = updateExpenseServiceRequestData.Description
	expense.Amount = updateExpenseServiceRequestData.Amount
	expense.Category = updateExpenseServiceRequestData.Category
	result, err = s.mongoDBRepo.Update(context.Background(), filter, expense)
	if !result.Success {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"UpdateExpenseResponse",
			updateExpenseServiceResponse{
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
		"UpdateExpenseResponse",
		updateExpenseServiceResponse{
			Message: "Operation is successful!",
			Success: true,
		},
	)
}

type removeExpenseServiceRequest struct {
	Action   	string `json:"action"`
	UserId		string `json:"userId"`
	ExpenseId	string `json:"expenseId"`
}

type removeExpenseServiceResponse struct {
	Message string 		   `json:"message"`
	Success bool 		   `json:"success"`
}

func (s *ExpenseService) HandleRemoveExpense(data []byte, replyTo string, correlationId string) { 
	removeExpenseServiceRequestData := removeExpenseServiceRequest{}
	err := json.Unmarshal(data, &removeExpenseServiceRequestData)
    if err != nil {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"RemoveExpenseResponse",
			removeExpenseServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
        return
    }
	
	filter := map[string]interface{}{
		"userId": removeExpenseServiceRequestData.UserId,
		"expenseId": removeExpenseServiceRequestData.ExpenseId,
	}
	result, err := s.mongoDBRepo.Find(context.Background(), filter)
	if !result.Success {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"RemoveExpenseResponse",
			removeExpenseServiceResponse{
				Message: "An error occured!",
				Success: false,
			},
		)
        return
	}
	if len(result.Data) == 0 {
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"RemoveExpenseResponse",
			removeExpenseServiceResponse{
				Message: "Expense not found!",
				Success: false,
			},
		)
		return
	}
	
	result, err = s.mongoDBRepo.Delete(context.Background(), filter)
	if !result.Success {
        log.Println(err)
		SendResponse(
			s.channel,
			replyTo,
			correlationId,
			"RemoveExpenseResponse",
			removeExpenseServiceResponse{
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
		"RemoveExpenseResponse",
		removeExpenseServiceResponse{
			Message: "Operation is successful!",
			Success: true,
		},
	)
}
