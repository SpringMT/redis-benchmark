package main

import (
	"github.com/go-redis/redis"
)

// RedisNewClient Redisクライアントを返す
// https://godoc.org/github.com/go-redis/redis#Options
func RedisNewClient(host string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		//MaxRetries:   1,
		//DialTimeout:  2,
		//ReadTimeout:  2,
		//WriteTimeout: 2,
	})

	//pong, err := client.Ping().Result()
	//if err != nil {
	//	fmt.Println(pong, err)
	//}
	// Output: PONG <nil>
	return client
}
