package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type loginWebRequest struct {
	Action   string `json:"action"`
	Email	 string `json:"email"`
	Password string `json:"password"`
}

type loginWebResponse struct {
	Message string 	`json:"message"`
	Success bool 	`json:"success"`
	Token  	string 	`json:"token"`
}

func HandleLoginRoute(ch *amqp.Channel) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        loginWebRequestData := &loginWebRequest{}
        err := json.NewDecoder(r.Body).Decode(&loginWebRequestData)
        if err != nil {
            http.Error(w, "Format is invalid!", http.StatusBadRequest)
            return
        }

        if loginWebRequestData.Email == "" || loginWebRequestData.Password == "" {
            http.Error(w, "\"Email\" and \"Password\" are required!", http.StatusBadRequest)
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

        SendRequest(ch, "userQueue", "Login", loginWebRequestData, replyQueue.Name, correlationId)

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
					token, tokenExists := data["token"].(string)
					if !tokenExists {
						http.Error(w, "\"Token\" field is not found!", http.StatusInternalServerError)
						return
					}

					loginWebResponseData := loginWebResponse{
						Message: "Login is successful!",
						Success: true,
						Token: token,
					}

					loginWebResponseDataJSON, err := json.Marshal(loginWebResponseData)
					if err != nil {
						http.Error(w, "Failed to convert response!", http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(loginWebResponseDataJSON)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}
		}
        http.Error(w, "No response is received!", http.StatusRequestTimeout)
    }
}

type registerWebRequest struct {
	Action   string `json:"action"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name 	 string `json:"name"`
}

type registerWebResponse struct {
	Message string 	`json:"message"`
	Success bool 	`json:"success"`
}

func HandleRegisterRoute(ch *amqp.Channel) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        registerWebRequestData := registerWebRequest{}
        err := json.NewDecoder(r.Body).Decode(&registerWebRequestData)
        if err != nil {
            http.Error(w, "Format is invalid!", http.StatusBadRequest)
            return
        }

        if registerWebRequestData.Email == "" || registerWebRequestData.Password == "" || registerWebRequestData.Name == "" {
            http.Error(w, "\"Email\", \"Password\" and \"Name\" are required", http.StatusBadRequest)
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

        SendRequest(ch, "userQueue", "Register", registerWebRequestData, replyQueue.Name, correlationId)

        for message := range messages {
			if message.CorrelationId == correlationId {
				serviceResponseData := serviceResponse{}
				err := json.Unmarshal(message.Body, &serviceResponseData)
				if err != nil {
					http.Error(w, "Failed to convert service response!", http.StatusInternalServerError)
					return
				}

				data := serviceResponseData.Data

				success, successExists := data["success"].(bool); 
				if !successExists {
					http.Error(w, "\"Success\" field is not found!", http.StatusInternalServerError)
					return
				}

				if success {
					registerWebResponseData := registerWebResponse{
						Message: "Register is successful!",
						Success: true,
					}
					
					registerWebResponseDataJSON, err := json.Marshal(registerWebResponseData)
					if err != nil {
						http.Error(w, "Failed to convert response!", http.StatusInternalServerError)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(registerWebResponseDataJSON)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
            }
        }

        http.Error(w, "No response is received!", http.StatusRequestTimeout)
    }
}
