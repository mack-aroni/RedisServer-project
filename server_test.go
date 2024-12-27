package main

import (
	"RedisServer-project/client"
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

func TestFooBar(t *testing.T) {
	in := map[string]string{
		"server":  "redis",
		"version": "6.0",
	}
	out := respWriteMap(in)
	fmt.Println(string(out))
}

func TestServerWithMultiCleints(t *testing.T) {
	server := NewServer(ServerConfig{})

	go func() {
		log.Fatal(server.Start())
	}()

	time.Sleep(time.Second)

	nClients := 10
	wg := sync.WaitGroup{}
	wg.Add(nClients)

	for i := 0; i < nClients; i++ {

		go func(x int) {
			c, err := client.NewClient("localhost:5001")
			if err != nil {
				log.Fatal(err)
			}

			defer c.Close()

			key := fmt.Sprintf("client_foo_%d", x)
			val := fmt.Sprintf("client_bar_%d", x)

			if err := c.Set(context.TODO(), key, val); err != nil {
				log.Fatal(err)
			}

			v, err := c.Get(context.TODO(), key)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("client %d got this val back %s\n", x, v)
			wg.Done()
		}(i)
	}

	wg.Wait()

	time.Sleep(time.Second)

	if len(server.peers) != 0 {
		t.Fatalf("expected 0 peers but got %d", len(server.peers))
	}

}
