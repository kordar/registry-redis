package registry_redis_test

import (
	"github.com/go-redis/redis"
	"github.com/kordar/registry"
	"github.com/kordar/registry-redis"
	"log"
	"testing"
	"time"
)

func TestRedisRegistry_Get(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "192.168.30.16:30202",
		Password: "940430Dex",
		DB:       1,
	})

	var redisnoderegistry registry.Registry = registry_redis.NewRedisNodeRegistry(client, &registry_redis.RedisNodeRegistryOptions{
		Prefix:  "ABC",
		Node:    "123.12.34.2:3320",
		Timeout: time.Second * 30,
		Channel: "BOB",
		Reload: func(value []string, channel string) {
			log.Println("--------------", value, channel)
		},
		Heartbeat: time.Second * 3,
	})
	redisnoderegistry.Listener()
	time.Sleep(5 * time.Second)
	_ = redisnoderegistry.Register()

	registry2 := registry_redis.NewRedisNodeRegistry(client, &registry_redis.RedisNodeRegistryOptions{
		Prefix:  "ABC",
		Node:    "123.12.34.3:3320",
		Timeout: time.Second * 30,
		Channel: "BOB",
		Reload: func(value []string, channel string) {
			log.Println("22222222", value, channel)
		},
		Heartbeat: time.Second * 3,
	})
	registry2.Listener()
	_ = registry2.Register()

	time.Sleep(100 * time.Second)
}
