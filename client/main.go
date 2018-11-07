package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	pb "github.com/abhinavdangeti/gRPCBench/protobuf"
	"github.com/jamiealquiza/tachymeter"
	"google.golang.org/grpc"
)

var RUN_TIME_SECS = flag.Int("runTimeSecs", 60, "int")
var NUM_MSGS_PER_SAMPLE = flag.Int("numMsgsPerSample", 10, "int")
var STREAM_CTX_TIMEOUT = flag.Int("streamCtxTimeout", 5, "int")

func init() {
	flag.Parse()
}

type client struct {
	name      string
	addresses []string
	conns     []*grpc.ClientConn
	t         *tachymeter.Tachymeter
}

func newClient(name string, addresses []string) (*client, error) {
	var conns []*grpc.ClientConn
	for _, address := range addresses {
		// Set up a connection to the server.
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("[%v] did not connect: %v", name, err)
		}
		conns = append(conns, conn)
	}

	return &client{
		name:      name,
		addresses: addresses,
		conns:     conns,
		t:         tachymeter.New(&tachymeter.Config{Size: 10000}),
	}, nil
}

func (c *client) run() {
	var wg sync.WaitGroup
	start := time.Now()
	for _, clientConn := range c.conns {
		wg.Add(1)
		go func(conn *grpc.ClientConn) {
			defer wg.Done()

			co := pb.NewEngageClient(conn)

			// Contact the server and print out its response.
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Duration(*STREAM_CTX_TIMEOUT)*time.Second)
			defer cancel()

			r, err := co.Greet(ctx, &pb.Greeting{Name: c.name, Msg: "hi"})
			if err != nil {
				log.Fatalf("[%v] could not greet: %v", c.name, err)
			} else if r.Msg != "hello "+c.name {
				log.Fatalf("[%v] unexpected response: %v", c.name, r.Msg)
			}

			stream, err := co.ShipData(ctx,
				&pb.Request{Name: c.name, Ask: int32(*NUM_MSGS_PER_SAMPLE)})
			if err != nil {
				log.Fatalf("[%v] could not request data to be shipped: %v", c.name, err)
			}
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil || resp == nil {
					log.Fatalf("[%v] could not receive data: %v", c.name, err)
				}
			}

			r, err = co.Greet(ctx, &pb.Greeting{Name: c.name, Msg: "bye"})
			if err != nil {
				log.Fatalf("[%v] could not complete: %v", c.name, err)
			} else if r.Msg != "goodbye "+c.name {
				log.Fatalf("[%v] unexpected response: %v", c.name, r.Msg)
			}
		}(clientConn)
	}
	wg.Wait()
	// Tachymeter is thread-safe.
	c.t.AddTime(time.Since(start))
}

func (c *client) terminate() {
	for _, conn := range c.conns {
		conn.Close()
	}
}

func (c *client) resultsStr() string {
	return fmt.Sprintf(
		"=======================================================\n"+
			"Client: %v\n"+
			"Addresses: %v\n"+
			"%v\n"+
			"=======================================================\n",
		c.name, c.addresses, c.t.Calc().String())
}

func dumpResultsToFile(clients []*client, dir string) error {
	timeNow := time.Now()
	newFile := strconv.FormatInt(timeNow.Unix(), 10) + ".results.log"
	file, err := os.Create(filepath.Join(dir, newFile))
	if err != nil {
		return err
	}

	fmt.Fprintf(file, timeNow.String()+"\n")
	fmt.Fprintf(file, "Run time (secs): "+strconv.Itoa(*RUN_TIME_SECS)+"\n")
	fmt.Fprintf(file, "Num messages streamed per sample: "+
		strconv.Itoa(*NUM_MSGS_PER_SAMPLE+2)+"\n")
	fmt.Fprintf(file, "\n")
	for _, client := range clients {
		fmt.Fprintf(file, client.resultsStr()+"\n")
	}
	return nil
}

func main() {
	var clients []*client
	for i, addresses := range [][]string{
		[]string{"127.0.0.1:12345", "127.0.0.1:23456"},
		[]string{"127.0.0.1:23456", "127.0.0.1:34567"},
		[]string{"127.0.0.1:34567", "127.0.0.1:12345"}} {
		client, err := newClient("client"+strconv.Itoa(i), addresses)
		if err != nil {
			log.Fatalf("Error setting up client: %v", err)
		}
		clients = append(clients, client)
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

		if time.Since(start) > time.Duration(*RUN_TIME_SECS)*time.Second {
			break
		}
	}

	err := dumpResultsToFile(clients, ".")
	if err != nil {
		log.Fatalf("Error dumping results to file: %v", err)
	}

	for _, cl := range clients {
		cl.terminate()
	}
}
