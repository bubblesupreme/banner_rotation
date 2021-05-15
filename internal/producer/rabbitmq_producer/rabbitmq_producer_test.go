package rabbitmqproducer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/bubblesupreme/banner_rotation/internal/producer"

	"github.com/NeowayLabs/wabbit"
	"github.com/NeowayLabs/wabbit/amqptest"
	"github.com/NeowayLabs/wabbit/amqptest/server"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type handleFunc func(deliveries <-chan wabbit.Delivery, done chan error)

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

	m := sync.Mutex{}
	handledClicks := make([]producer.Action, 0, nAction)
	handledShows := make([]producer.Action, 0, nAction)

	fakeServer := server.NewServer(url)
	fakeServer.Start()

	conn, err := amqptest.Dial(url)
	assert.NoError(t, err)

	p, err := NewProducer(conn, exchangeName, clickRoutingKey, showRoutingKey)
	defer func() {
		assert.NoError(t, p.Shutdown())
	}()
	assert.NoError(t, err)

	cClick, err := createConsumer(url, exchangeName, clickQueueName, clickRoutingKey, "", func(deliveries <-chan wabbit.Delivery, done chan error) {
		for d := range deliveries {
			a, err := byteArrayToAction(d.Body())
			assert.NoError(t, err)
			m.Lock()
			handledClicks = append(handledClicks, a)
			m.Unlock()
			d.Ack(false)
		}
		log.Printf("handle: deliveries channel closed")
		done <- nil
	})
	assert.NoError(t, err)

	cShow, err := createConsumer(url, exchangeName, showQueueName, showRoutingKey, "", func(deliveries <-chan wabbit.Delivery, done chan error) {
		for d := range deliveries {
			a, err := byteArrayToAction(d.Body())
			assert.NoError(t, err)
			m.Lock()
			handledShows = append(handledShows, a)
			m.Unlock()
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
			assert.NoError(t, p.Click(producer.Action{
				SlotID:   a.SlotID,
				BannerID: a.BannerID,
				GroupID:  a.GroupID,
			}))
		}
	}()

	go func() {
		defer wg.Done()

		for _, a := range shows {
			assert.NoError(t, p.Show(producer.Action{
				SlotID:   a.SlotID,
				BannerID: a.BannerID,
				GroupID:  a.GroupID,
			}))
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
				m.Lock()
				if len(handledClicks) == nAction && len(handledShows) == nAction {
					// check clicks
					checkHandledActions(t, clicks, handledClicks)
					// check shows
					checkHandledActions(t, shows, handledShows)
					return
				}
				m.Unlock()
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
	conn    *amqptest.Conn
	channel wabbit.Channel
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

	fakeServer := server.NewServer(amqpURI)
	fakeServer.Start()

	var err error
	c.conn, err = amqptest.Dial(amqpURI)
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
		wabbit.Option{
			"durable":    true,
			"autoDelete": false,
			"exclusive":  false,
			"noWait":     false,
		}); err != nil {
		return consumer{}, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	q, err := c.channel.QueueDeclare(
		queueName, wabbit.Option{
			"durable":          false,
			"deleteWhenUnused": false,
			"exclusive":        true,
			"noWait":           false,
		},
	)
	if err != nil {
		return consumer{}, fmt.Errorf("failed to declare a queue for click action: %w", err)
	}
	if err = c.channel.QueueBind(
		q.Name(),
		routingKey,
		exchangeName,
		wabbit.Option{},
	); err != nil {
		return consumer{}, fmt.Errorf("failed to bind a queue for click action: %w", err)
	}

	deliveries, err := c.channel.Consume(
		q.Name(), // name
		tag,      // consumerTag,
		wabbit.Option{
			"noAck":     false,
			"exclusive": false,
			"noLocal":   false,
			"noWait":    false,
		},
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
