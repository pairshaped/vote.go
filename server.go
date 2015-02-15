package main

import (
	"flag"
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

var (
	redisAddress   = flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
)

func main() {
	flag.Parse()

	redisPool := redis.NewPool(func() (redis.Conn, error) {
		redisClient, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			return nil, err
		}

		return redisClient, err
	}, *maxConnections)

	defer redisPool.Close()

	m := martini.Classic()

	m.Map(redisPool)
	m.Use(render.Renderer())

	m.Get("/vote/:id", func(redisPool *redis.Pool, params martini.Params, r render.Render) {

		if len(params["id"]) < 8 {
			key := "votes/" + params["id"]

			redisClient := redisPool.Get()
			defer redisClient.Close()

			votes, err := redisClient.Do("INCR", key)

			if err != nil {
				message := fmt.Sprintf("Unable to vote for: %s", key)
				r.JSON(400, map[string]interface{}{
					"status":  "ERROR",
					"message": message})

			} else {
				message := fmt.Sprintf("Voted for: %s, votes at: %d", key, votes)
				r.JSON(200, map[string]interface{}{
					"status":  "SUCCESS",
					"message": message})
			}

		} else {
			r.JSON(400, map[string]interface{}{
				"status":  "ERROR",
				"message": "Not a valid vote"})
		}
	})

	m.Run()
}
