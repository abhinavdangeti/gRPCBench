//
// gRPC: Benchmarks
// Author: Abhinav Dangeti
//

package main

import (
	"context"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	pb "github.com/abhinavdangeti/gRPCBench/protobuf"
	"github.com/jamiealquiza/tachymeter"
	"google.golang.org/grpc"
)

type client struct {
	name      string
	addresses []string
	t         *tachymeter.Tachymeter
}

func newClient(name string, addresses []string) *client {
	return &client{
		name:      name,
		addresses: addresses,
		t:         tachymeter.New(&tachymeter.Config{Size: 500}),
	}
}

func (c *client) run() {
	var wg sync.WaitGroup
	start := time.Now()
	for _, addr := range c.addresses {
		wg.Add(1)
		go func(address string) {
			defer wg.Done()

			name := c.name + "_" + address

			// Set up a connection to the server.
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("[%v] did not connect: %v", name, err)
			}
			defer conn.Close()
			co := pb.NewEngageClient(conn)

			// Contact the server and print out its response.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			r, err := co.Greet(ctx, &pb.Greeting{Name: name, Msg: "hi"})
			if err != nil {
				log.Fatalf("[%v] could not greet: %v", name, err)
			} else if r.Msg != "hello " + name {
				log.Fatalf("[%v] unexpected response: %v", name, r.Msg)
			}

			stream, err := co.ShipData(ctx, &pb.Request{Name: name, Ask: 10})
			if err != nil {
				log.Fatalf("[%v] could not request data to be shipped: %v", name, err)
			}
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil || resp == nil{
					log.Fatalf("[%v] could not receive data: %v", name, err)
				}
			}

			r, err = co.Greet(ctx, &pb.Greeting{Name: name, Msg: "bye"})
			if err != nil {
				log.Fatalf("[%v] could not complete: %v", name, err)
			} else if r.Msg != "goodbye " + name {
				log.Fatalf("[%v] unexpected response: %v", name, r.Msg)
			}
		}(addr)
	}
	wg.Wait()
	// Tachymeter is thread-safe.
	c.t.AddTime(time.Since(start))
}

func (c *client) results() {
	log.Printf("\n===================================\n" +
		"Client: %v\n" +
		"Addresses: %v\n" +
		"%v\n" +
		"===================================\n",
		c.name, c.addresses, c.t.Calc().String())
}

func main() {
	var clients []*client
	for i, addresses := range [][]string{
		[]string{"127.0.0.1:12345", "127.0.0.1:23456"},
		[]string{"127.0.0.1:23456", "127.0.0.1:34567"},
		[]string{"127.0.0.1:34567", "127.0.0.1:12345"}} {
		clients = append(clients, newClient("client"+strconv.Itoa(i), addresses))
	}

	start := time.Now()
	for {
		var wg sync.WaitGroup
		for _, cl := range clients {
			wg.Add(1)
			go func(c *client) {
				c.run()
				wg.Done()
			}(cl)
		}
		wg.Wait()

		if time.Since(start) > 20*time.Second {
			break
		}
	}

	for _, cl := range clients {
		cl.results()
	}
}
