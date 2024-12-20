package services

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type response struct {
    Action string      `json:"action"`
    Data   interface{} `json:"data"`
}

func SendResponse(ch *amqp.Channel, replyTo string, correlationID string, action string, data interface{}) {
    responseData := response{
        Action: action,
        Data: data,
    }

    responseDataJSON, err := json.Marshal(responseData)
    if err != nil {
        log.Println("Failed to encode response data to JSON: %v", err)
        return
    }

    err = ch.Publish(
        "",
        replyTo,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            CorrelationId: correlationID,
            Body: responseDataJSON,
            Headers: amqp.Table{
                "action": action,
            },
        },
    )
    if err != nil {
        log.Println("Failed to publish a response message: %v", err)
    }
}