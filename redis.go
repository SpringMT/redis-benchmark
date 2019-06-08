package main

import (
	"github.com/go-redis/redis"
)

type RedisClient struct {
	Client *redis.Client
}

func (rc *RedisClient) increment(key string) (int64, error) {
	pipe := rc.Client.Pipeline()
	incr := pipe.Incr(key)
	_, err := pipe.Exec()
	return incr.Val(), err
}

func (rc *RedisClient) close() {
	rc.Client.Close()
}

// RedisNewClient Redisクライアントを返す
// https://godoc.org/github.com/go-redis/redis#Options
func RedisNewClient(host string) RedisClient {
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
	return RedisClient{Client: client}
}
