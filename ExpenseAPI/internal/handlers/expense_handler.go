package handlers

import (
	"expense/internal/middlewares"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type getExpenseRequest struct {
	Action   string 				`json:"action"`
	UserId	 string 				`json:"userId"`
	Filter   map[string]interface{} `json:"filter"`
}

type getExpenseResponse struct {
	Message  string 		  `json:"message"`
	Success  bool 			  `json:"success"`
	Expenses interface{}	  `json:"expense"`
}

func HandleGetExpenseRoute(ch *amqp.Channel) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
		getExpenseRequestData := &getExpenseRequest{}
		getExpenseRequestData.Filter = make(map[string]interface{})

		query := r.URL.Query()
		for key, _ := range query {
			if key != "expenseId" && key != "category" {
				http.Error(w, "Invalid query parameter!", http.StatusBadRequest)
				return
			}
		}
		expenseId := query.Get("expenseId")
		category := query.Get("category")
		if expenseId != "" && category != "" {
			http.Error(w, "At most one of the \"ExpenseId\" or \"Category\" fields can be used!", http.StatusBadRequest)
			return
		} else if expenseId != "" && category == "" {
			getExpenseRequestData.Filter["expenseId"] = expenseId
		} else if expenseId == "" && category != "" {
			getExpenseRequestData.Filter["category"] = category
		}

        correlationId := uuid.New().String()

		replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
        if err != nil {
            http.Error(w, "Failed to create queue!", http.StatusInternalServerError)
            return
        }

        messages, err := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)
        if err != nil {
            http.Error(w, "Failed to create consumer!", http.StatusInternalServerError)
            return
        }

		userId, err := middlewares.GetUserIdFromRequest(r)
		if err != nil {
			http.Error(w, "Failed to get user id!", http.StatusInternalServerError)
			return
		}
		
		getExpenseRequestData.UserId = userId
        SendRequest(ch, "expenseQueue", "GetExpense", getExpenseRequestData, replyQueue.Name, correlationId)

        for message := range messages {
			if message.CorrelationId == correlationId {
				serviceResponseData := serviceResponse{}
				err := json.Unmarshal(message.Body, &serviceResponseData)
				if err != nil {
					http.Error(w, "Failed to convert service response!", http.StatusInternalServerError)
					return
				}

				data := serviceResponseData.Data

				success, successExists := data["success"].(bool)
				if !successExists {
					http.Error(w, "\"Success\" field is not found!", http.StatusInternalServerError)
					return
				}

				if success {

					expenses := make([]map[string]interface{}, 0)
					for _, eachData := range data["expenses"].([]interface{}) {
						expense := eachData.(map[string]interface{})
						expenses = append(expenses, expense)
					}

					getExpenseResponseData := getExpenseResponse{
						Message: "Operation is successful!",
						Success: true,
						Expenses: expenses,
					}

					getExpenseResponseDataJSON, err := json.Marshal(getExpenseResponseData)
					if err != nil {
						http.Error(w, "Failed to convert response!", http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(getExpenseResponseDataJSON)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}
		}
        http.Error(w, "No response is received!", http.StatusRequestTimeout)
    }
}

type addExpenseRequest struct {
	Action   	string 	`json:"action"`
	UserId		string 	`json:"userId"`
	Description string 	`json:"description"`
	Amount 		float32 `json:"amount"`
	Category 	string 	`json:"category"`
}

type addExpenseResponse struct {
	Message string 		   `json:"message"`
	Success bool 		   `json:"success"`
}

func HandleAddExpenseRoute(ch *amqp.Channel) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        addExpenseRequestData := &addExpenseRequest{}
        err := json.NewDecoder(r.Body).Decode(&addExpenseRequestData)
        if err != nil {
            http.Error(w, "Format is invalid!", http.StatusBadRequest)
            return
        }

        if addExpenseRequestData.Description == "" || addExpenseRequestData.Amount == 0 || addExpenseRequestData.Category == "" {
            http.Error(w, "\"Description\", \"Amount\" and \"Category\" are required!", http.StatusBadRequest)
            return
        }

        correlationId := uuid.New().String()

		replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
        if err != nil {
            http.Error(w, "Failed to create queue!", http.StatusInternalServerError)
            return
        }

        messages, err := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)
        if err != nil {
            http.Error(w, "Failed to create consumer!", http.StatusInternalServerError)
            return
        }

		userId, err := middlewares.GetUserIdFromRequest(r)
		if err != nil {
			http.Error(w, "Failed to get user id!", http.StatusInternalServerError)
			return
		}
		
		addExpenseRequestData.UserId = userId
        SendRequest(ch, "expenseQueue", "AddExpense", addExpenseRequestData, replyQueue.Name, correlationId)

        for message := range messages {
			if message.CorrelationId == correlationId {
				serviceResponseData := serviceResponse{}
				err := json.Unmarshal(message.Body, &serviceResponseData)
				if err != nil {
					http.Error(w, "Failed to convert service response!", http.StatusInternalServerError)
					return
				}

				data := serviceResponseData.Data

				success, successExists := data["success"].(bool)
				if !successExists {
					http.Error(w, "\"Success\" field is not found!", http.StatusInternalServerError)
					return
				}

				if success {

					addExpenseResponseData := addExpenseResponse{
						Message: "Operation is successful!",
						Success: true,
					}

					addExpenseResponseDataJSON, err := json.Marshal(addExpenseResponseData)
					if err != nil {
						http.Error(w, "Failed to convert response!", http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(addExpenseResponseDataJSON)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}
		}
        http.Error(w, "No response is received!", http.StatusRequestTimeout)
    }
}

type updateExpenseRequest struct {
	Action   	string 	`json:"action"`
	UserId		string  `json:"userId"`
	ExpenseId	string 	`json:"expenseId"`
	Description string 	`json:"description"`
	Amount 		float32 `json:"amount"`
	Category 	string 	`json:"category"`
}

type updateExpenseResponse struct {
	Message string 		   `json:"message"`
	Success bool 		   `json:"success"`
}

func HandleUpdateExpenseRoute(ch *amqp.Channel) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        updateExpenseRequestData := &updateExpenseRequest{}
        err := json.NewDecoder(r.Body).Decode(&updateExpenseRequestData)
        if err != nil {
            http.Error(w, "Format is invalid!", http.StatusBadRequest)
            return
        }

        if updateExpenseRequestData.ExpenseId == "" || updateExpenseRequestData.Description == "" || updateExpenseRequestData.Amount == 0 || updateExpenseRequestData.Category == "" {
            http.Error(w, "\"ExpenseId\", \"Description\", \"Amount\" and \"Category\" are required!", http.StatusBadRequest)
            return
        }

        correlationId := uuid.New().String()

		replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
        if err != nil {
            http.Error(w, "Failed to create queue!", http.StatusInternalServerError)
            return
        }

        messages, err := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)
        if err != nil {
            http.Error(w, "Failed to create consumer!", http.StatusInternalServerError)
            return
        }

		userId, err := middlewares.GetUserIdFromRequest(r)
		if err != nil {
			http.Error(w, "Failed to get user id!", http.StatusInternalServerError)
			return
		}
		
		updateExpenseRequestData.UserId = userId
        SendRequest(ch, "expenseQueue", "UpdateExpense", updateExpenseRequestData, replyQueue.Name, correlationId)

        for message := range messages {
			if message.CorrelationId == correlationId {
				serviceResponseData := serviceResponse{}
				err := json.Unmarshal(message.Body, &serviceResponseData)
				if err != nil {
					http.Error(w, "Failed to convert service response!", http.StatusInternalServerError)
					return
				}

				data := serviceResponseData.Data

				success, successExists := data["success"].(bool)
				if !successExists {
					http.Error(w, "\"Success\" field is not found!", http.StatusInternalServerError)
					return
				}

				if success {

					updateExpenseResponseData := updateExpenseResponse{
						Message: "Operation is successful!",
						Success: true,
					}

					updateExpenseResponseDataJSON, err := json.Marshal(updateExpenseResponseData)
					if err != nil {
						http.Error(w, "Failed to convert response!", http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(updateExpenseResponseDataJSON)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}
		}
        http.Error(w, "No response is received!", http.StatusRequestTimeout)
    }
}

type removeExpenseRequest struct {
	Action   	string `json:"action"`
	UserId		string `json:"userId"`
	ExpenseId	string `json:"expenseId"`
}

type removeExpenseResponse struct {
	Message string 		   `json:"message"`
	Success bool 		   `json:"success"`
}

func HandleRemoveExpenseRoute(ch *amqp.Channel) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        removeExpenseRequestData := &removeExpenseRequest{}
        err := json.NewDecoder(r.Body).Decode(&removeExpenseRequestData)
        if err != nil {
            http.Error(w, "Format is invalid!", http.StatusBadRequest)
            return
        }

        if removeExpenseRequestData.ExpenseId == "" {
            http.Error(w, "\"ExpenseId\" is required!", http.StatusBadRequest)
            return
        }

        correlationId := uuid.New().String()

		replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
        if err != nil {
            http.Error(w, "Failed to create queue!", http.StatusInternalServerError)
            return
        }

        messages, err := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)
        if err != nil {
            http.Error(w, "Failed to create consumer!", http.StatusInternalServerError)
            return
        }

		userId, err := middlewares.GetUserIdFromRequest(r)
		if err != nil {
			http.Error(w, "Failed to get user id!", http.StatusInternalServerError)
			return
		}
		
		removeExpenseRequestData.UserId = userId
        SendRequest(ch, "expenseQueue", "RemoveExpense", removeExpenseRequestData, replyQueue.Name, correlationId)

        for message := range messages {
			if message.CorrelationId == correlationId {
				serviceResponseData := serviceResponse{}
				err := json.Unmarshal(message.Body, &serviceResponseData)
				if err != nil {
					http.Error(w, "Failed to convert service response!", http.StatusInternalServerError)
					return
				}

				data := serviceResponseData.Data

				success, successExists := data["success"].(bool)
				if !successExists {
					http.Error(w, "\"Success\" field is not found!", http.StatusInternalServerError)
					return
				}

				if success {

					removeExpenseResponseData := removeExpenseResponse{
						Message: "Operation is successful!",
						Success: true,
					}

					removeExpenseResponseDataJSON, err := json.Marshal(removeExpenseResponseData)
					if err != nil {
						http.Error(w, "Failed to convert response!", http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(removeExpenseResponseDataJSON)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}
		}
        http.Error(w, "No response is received!", http.StatusRequestTimeout)
    }
}
