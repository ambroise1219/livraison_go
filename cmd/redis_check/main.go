package main

import (
	"fmt"
	"github.com/ambroise1219/livraison_go/config"
)

func main() {
	client := config.GetRedisClient()
	if client == nil {
		fmt.Println("redis: client nil")
		return
	}
	if err := client.Ping(client.Context()).Err(); err != nil {
		fmt.Println("redis: ping error:", err)
		return
	}
	fmt.Println("redis: OK")
}
