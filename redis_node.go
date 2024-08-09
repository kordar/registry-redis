package registry_redis

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

type RedisNodeRegistryOptions struct {
	Prefix    string
	Node      string
	Timeout   time.Duration
	Channel   string
	Heartbeat time.Duration
	Reload    func(data []string, channel string)
}

// RedisNodeRegistry redis节点注册
type RedisNodeRegistry struct {
	options   *RedisNodeRegistryOptions
	once      sync.Once
	heartOnce sync.Once
	client    *redis.Client
}

func NewRedisNodeRegistry(client *redis.Client, options *RedisNodeRegistryOptions) *RedisNodeRegistry {
	return &RedisNodeRegistry{
		client:    client,
		once:      sync.Once{},
		heartOnce: sync.Once{},
		options:   options,
	}
}

func (r *RedisNodeRegistry) key() string {
	return fmt.Sprintf("%s:%s", r.options.Prefix, r.options.Node)
}

func (r *RedisNodeRegistry) Get() (interface{}, error) {
	keys := r.client.Keys(r.options.Prefix + ":*")
	cmd := r.client.MGet(keys.Val()...)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

// Remove 移除配置
func (r *RedisNodeRegistry) Remove() error {
	if err := r.client.Del(r.key()).Err(); err == nil {
		return PubMessage(r.client, r.options.Channel, "reload")
	} else {
		return err
	}
}

// Register 将节点信息写入到redis中,并向订阅者进行通知
func (r *RedisNodeRegistry) Register() error {
	r.heartOnce.Do(func() {
		ticker := time.NewTicker(r.options.Heartbeat)
		go func() {
			for {
				select {
				case <-ticker.C:
					fmt.Println("heartbeat run.....")
					r.reload(r.options.Channel)
					break
				}
			}
		}()
	})
	data := map[string]string{
		"node":         r.options.Node,
		"refresh_time": time.Now().Format("2006-01-02 15:04:05"),
		"status":       "online",
	}
	marshal, _ := json.Marshal(&data)
	if err := r.client.Set(r.key(), string(marshal), r.options.Timeout).Err(); err == nil {
		return PubMessage(r.client, r.options.Channel, "reload")
	} else {
		return err
	}
}

func (r *RedisNodeRegistry) Listener() {
	r.once.Do(func() {
		go SubMessage(r.client, Event{
			Channel: r.options.Channel,
			Fn: func(payload string, channel string) {
				if payload == "reload" {
					r.reload(channel)
				}
			},
		})
	})
}

func (r *RedisNodeRegistry) reload(channel string) {
	data := make([]string, 0)
	if val, err := r.Get(); err == nil {
		vv := val.([]interface{})
		for _, v := range vv {
			data = append(data, v.(string))
		}
	}
	if r.options.Reload != nil {
		r.options.Reload(data, channel)
	}
}
