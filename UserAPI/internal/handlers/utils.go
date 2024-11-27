package handlers

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

func SendRequest(ch *amqp.Channel, queueName string, action string, data interface{}, replyTo string, corrId string) {

	dataJSON, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}
	err = ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			ReplyTo: replyTo,
			CorrelationId: corrId,
			Body: dataJSON,
			Headers: amqp.Table{
				"action": action,
			},
		},
	)
	if err != nil {
		log.Println(err)
	}
}

type serviceResponse struct {
	Action  string `json:"action"`
	Data    map[string]interface{} `json:"data"`
}
