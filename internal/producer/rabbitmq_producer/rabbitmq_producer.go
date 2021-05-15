package rabbitmqproducer

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bubblesupreme/banner_rotation/internal/producer"

	"github.com/NeowayLabs/wabbit"

	log "github.com/sirupsen/logrus"
)

type publisher struct {
	connection      wabbit.Conn
	channel         wabbit.Channel
	clickRoutingKey string
	showRoutingKey  string
	exchangeName    string
}

func NewProducer(conn wabbit.Conn, exchangeName, clickRoutingKey, showRoutingKey string) (producer.Producer, error) {
	log.Info("got connection to RabbitMQ")

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel from connection: %w", err)
	}

	log.Info("open a channel from connection")
	if err := channel.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	return &publisher{
		connection:      conn,
		channel:         channel,
		clickRoutingKey: clickRoutingKey,
		showRoutingKey:  showRoutingKey,
		exchangeName:    exchangeName,
	}, nil
}

func (p *publisher) Shutdown() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("failed to cancel consumer: %w", err)
	}

	if err := p.connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	defer log.Info("AMQP producer shutdown OK")

	return nil
}

func (p *publisher) Show(a producer.Action) error {
	return p.publish(a, p.showRoutingKey)
}

func (p *publisher) Click(a producer.Action) error {
	return p.publish(a, p.clickRoutingKey)
}

func (p *publisher) publish(a producer.Action, routingKey string) error {
	b, err := actionToByteArray(a)
	if err != nil {
		return err
	}

	err = p.channel.Publish(
		p.exchangeName, // publish to an exchange
		routingKey,     // routing to 0 or more queues
		b,
		wabbit.Option{
			"deliveryMode": 2,
			"contentType":  "text/plain",
		},
	)

	if err == nil {
		log.WithFields(log.Fields{
			"slot id":     a.SlotID,
			"banner id":   a.BannerID,
			"group id":    a.GroupID,
			"routing key": routingKey,
		}).Info("publish an action")
	}

	return err
}

func actionToByteArray(a producer.Action) ([]byte, error) {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(&a)
	return b.Bytes(), err
}
