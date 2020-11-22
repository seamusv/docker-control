package main

import (
	"fmt"
	"github.com/docker/docker/client"
)

const (
	redisContainer = "/local_redis_1"
)

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Println("Unable to create docker client")
		panic(err)
	}

	redisCli, err := NewRedisCli(cli, redisContainer)
	if err != nil {
		panic(err)
	}

	keys, err := redisCli.Keys("*")
	if err != nil {
		panic(err)
	}
	for i, key := range keys {
		fmt.Printf("%d) %s\n", i+1, key)
	}

	votes, err := redisCli.Get("choice.00000000-0000-0000-0000-000000000002.votes")
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nVotes: %s\n", votes)
}
