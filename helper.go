package registry_redis

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
)

type Event struct {
	Channel string
	Fn      func(payload string, channel string)
}

func PubMessage(c *redis.Client, channel, msg string) error {
	cmd := c.Publish(channel, msg)
	return cmd.Err()
}

func SubMessage(c *redis.Client, events ...Event) error {
	if events == nil || len(events) == 0 {
		return errors.New("subscribe channel fail")
	}

	fn := map[string]func(payload string, channel string){}
	channels := make([]string, 0)
	for _, event := range events {
		channels = append(channels, event.Channel)
		fn[event.Channel] = event.Fn
	}

	pubsub := c.Subscribe(channels...)
	_, err := pubsub.Receive()
	if err != nil {
		return err
	}

	ch := pubsub.Channel()
	for msg := range ch {
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!")
		fn[msg.Channel](msg.Payload, msg.Channel)
	}

	return nil
}
