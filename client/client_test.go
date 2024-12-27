package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestNewClientRedisClient(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:5001",
		Password: "",
		DB:       0,
	})
	fmt.Println(rdb)
	fmt.Println("this is working")

	if err := rdb.Set(context.Background(), "key", "value", 0).Err(); err != nil {
		panic(err)
	}

	// val, err := rdb.Get(context.TODO(), "key")
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(val)
}

func TestNewClients(t *testing.T) {
	nClients := 10
	wg := sync.WaitGroup{}
	wg.Add(nClients)

	for i := 0; i < nClients; i++ {

		go func(x int) {
			c, err := NewClient("localhost:5001")
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

			fmt.Printf("client %d got this val back: %s\n", x, v)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestNewClient1(t *testing.T) {
	c, err := NewClient("localhost:5001")
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		fmt.Println("SET =>", fmt.Sprintf("bar_%d", i))

		if err := c.Set(context.TODO(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
			log.Fatal(err)
		}

		val, err := c.Get(context.TODO(), fmt.Sprintf("foo_%d", i))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("GET =>", val)
	}
}
