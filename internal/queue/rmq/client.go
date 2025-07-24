package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ilivestrong/email_warmup_service/internal/queue/events"
	"github.com/streadway/amqp"
)

const defaultQueue = "send_email"

type (
	SendEmailEvent        = events.SendEmailEvent
	SendEmailEventHandler = events.SendEmailEventHandler

	Client struct {
		conn    *amqp.Connection
		channel *amqp.Channel
	}
)

func New(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	_, err = ch.QueueDeclare(defaultQueue, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Client{conn: conn, channel: ch}, nil
}

func (c *Client) Publish(ctx context.Context, event *SendEmailEvent) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return c.channel.Publish("", defaultQueue, false, false, amqp.Publishing{ContentType: "application/json", Body: b})
}

func (c *Client) Consume(ctx context.Context, handler SendEmailEventHandler) error {
	msgs, err := c.channel.Consume(defaultQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return c.Close()
		case msg := <-msgs:
			e := new(SendEmailEvent)
			if err := json.Unmarshal(msg.Body, e); err != nil {
				log.Printf("invalid email event: %v", err)
				msg.Ack(false)
				continue
			}
			fmt.Printf("\n[NEW EVENT]: email: %s\n", e.ToAddress)
			err := handler(ctx, e)
			if err != nil {
				fmt.Println("failed to process event")
				msg.Nack(false, false)
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (c *Client) Close() error {
	time.Sleep(100 * time.Millisecond)
	c.channel.Close()
	return c.conn.Close()
}
