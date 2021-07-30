package main

import "github.com/go-redis/redis/v8"


func startRedis() *redis.Client{
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisUrl,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return client
}

type Message struct {
	Event map[string]string
	Data map[string]string

}