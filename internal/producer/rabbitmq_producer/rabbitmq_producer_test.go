package rabbitmqproducer

import (
	"banner_rotation/internal/producer"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"github.com/stretchr/testify/assert"
)

type handleFunc func(deliveries <-chan amqp.Delivery, done chan error)

func TestProducer(t *testing.T) {
	url := "amqp://guest:guest@localhost:5672/"
	exchangeName := "test"
	clickQueueName := "click"
	showQueueName := "show"
	clickRoutingKey := "click_key"
	showRoutingKey := "show_key"
	nAction := 5
	maxID := 5
	clicks := make([]producer.Action, nAction)
	for i := 0; i < nAction; i++ {
		clicks[i].BannerID = rand.Intn(maxID)
		clicks[i].SlotID = rand.Intn(maxID)
		clicks[i].GroupID = rand.Intn(maxID)
	}
	shows := make([]producer.Action, nAction)
	for i := 0; i < nAction; i++ {
		shows[i].BannerID = rand.Intn(maxID)
		shows[i].SlotID = rand.Intn(maxID)
		shows[i].GroupID = rand.Intn(maxID)
	}
	handledClicks := make([]producer.Action, 0, nAction)
	handledShows := make([]producer.Action, 0, nAction)

	p, err := NewProducer(url, exchangeName, clickRoutingKey, showRoutingKey)
	defer func() {
		assert.NoError(t, p.Shutdown())
	}()
	assert.NoError(t, err)

	cClick, err := createConsumer(url, exchangeName, clickQueueName, clickRoutingKey, "", func(deliveries <-chan amqp.Delivery, done chan error) {
		for d := range deliveries {
			a, err := byteArrayToAction(d.Body)
			assert.NoError(t, err)
			handledClicks = append(handledClicks, a)
			d.Ack(false)
		}
		log.Printf("handle: deliveries channel closed")
		done <- nil
	})
	assert.NoError(t, err)

	cShow, err := createConsumer(url, exchangeName, showQueueName, showRoutingKey, "", func(deliveries <-chan amqp.Delivery, done chan error) {
		for d := range deliveries {
			a, err := byteArrayToAction(d.Body)
			assert.NoError(t, err)
			handledShows = append(handledShows, a)
			d.Ack(false)
		}
		log.Printf("handle: deliveries channel closed")
		done <- nil
	})
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		for _, a := range clicks {
			err = p.Click(producer.Action{
				SlotID:   a.SlotID,
				BannerID: a.BannerID,
				GroupID:  a.GroupID,
			})
			assert.NoError(t, err)
		}
	}()

	go func() {
		defer wg.Done()

		for _, a := range shows {
			err = p.Show(producer.Action{
				SlotID:   a.SlotID,
				BannerID: a.BannerID,
				GroupID:  a.GroupID,
			})
			assert.NoError(t, err)
		}
	}()

	finishByTime := false
	go func() {
		defer wg.Done()
		defer func() {
			assert.NoError(t, cShow.shutdown())
		}()
		defer func() {
			assert.NoError(t, cClick.shutdown())
		}()

		timeout := time.After(10 * time.Second)

		for {
			select {
			case <-timeout:
				finishByTime = true
				return
			default:
				if len(handledClicks) == nAction && len(handledShows) == nAction {
					// check clicks
					checkHandledActions(t, clicks, handledClicks)
					// check shows
					checkHandledActions(t, shows, handledShows)
					return
				}
			}
		}
	}()

	wg.Wait()
	if finishByTime {
		t.Fatal("test didn't finish in time")
	}
	assert.True(t, nAction > 0)
}

type consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func createConsumer(amqpURI, exchangeName, queueName, routingKey, tag string, fn handleFunc) (consumer, error) {
	c := consumer{
		conn:    nil,
		channel: nil,
		tag:     tag,
		done:    make(chan error),
	}

	var err error
	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return c, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return consumer{}, fmt.Errorf("failed to open a channel from connection: %w", err)
	}

	log.Info("open a channel from connection")
	if err := c.channel.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		return consumer{}, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	q, err := c.channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		true,      // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return consumer{}, fmt.Errorf("failed to declare a queue for click action: %w", err)
	}
	if err = c.channel.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	); err != nil {
		return consumer{}, fmt.Errorf("failed to bind a queue for click action: %w", err)
	}

	deliveries, err := c.channel.Consume(
		q.Name, // name
		tag,    // consumerTag,
		false,  // noAck
		false,  // exclusive
		false,  // noLocal
		false,  // noWait
		nil,    // arguments
	)
	if err != nil {
		return c, fmt.Errorf("gueue consume failed: %w", err)
	}
	go fn(deliveries, c.done)

	return c, nil
}

func (c *consumer) shutdown() error {
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("failed to cancel consumer: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	defer log.Info("AMQP shutdown OK")

	return <-c.done
}

func byteArrayToAction(b []byte) (producer.Action, error) {
	w := bytes.NewReader(b)
	a := producer.Action{}
	err := json.NewDecoder(w).Decode(&a)
	return a, err
}

func checkHandledActions(t *testing.T, actions []producer.Action, handled []producer.Action) {
	assert.True(t, len(actions) == len(handled))

	isHandled := make([]bool, len(actions))
	for _, h := range handled {
		for i, c := range actions {
			if h.SlotID == c.SlotID &&
				h.BannerID == c.BannerID &&
				h.GroupID == c.GroupID &&
				isHandled[i] == false {
				isHandled[i] = true
				continue
			}
		}
	}
	// check all were handled
	for _, h := range isHandled {
		assert.True(t, h)
	}
}
