package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"

	pb "github.com/abhinavdangeti/gRPCBench/protobuf"
	"github.com/jamiealquiza/tachymeter"
	"google.golang.org/grpc"
)

var RUN_TIME_SECS = flag.Int("runTimeSecs", 60, "int")
var CONN_CONCURRENCY = flag.Int("connectionConcurrency", 1, "int")
var NUM_MSGS_PER_SAMPLE = flag.Int("numMsgsPerSample", 10, "int")
var VARY_NUM_MSGS = flag.Bool("varyNumMsgs", false, "bool: vary number of messsages streamed")
var STREAM_CTX_TIMEOUT = flag.Int("streamCtxTimeout", 5, "int")
var OPTION = flag.String("option", "shipbulkdata", "string: greet/shipdata/shipbulkdata")
var CPU_PROFILE = flag.String("cpuprofile", "", "write cpu profile to `file`")

func init() {
	flag.Parse()
	if *CONN_CONCURRENCY < 1 {
		*CONN_CONCURRENCY = 1
	}
}

type client struct {
	name      string
	addresses []string
	conns     []*grpc.ClientConn
	tachs     []*tachymeter.Tachymeter
}

func newClient(name string, addresses []string) (*client, error) {
	var conns []*grpc.ClientConn
	var tachs []*tachymeter.Tachymeter
	for _, address := range addresses {
		// Set up a connection to the server.
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("[%v] did not connect: %v", name, err)
		}
		conns = append(conns, conn)
		tachs = append(tachs, tachymeter.New(&tachymeter.Config{Size: 10000}))
	}

	return &client{
		name:      name,
		addresses: addresses,
		conns:     conns,
		tachs:     tachs,
	}, nil
}

func (c *client) run() {
	var wg sync.WaitGroup
	for k, clientConn := range c.conns {
		for i := 0; i < (*CONN_CONCURRENCY); i++ {
			wg.Add(1)
			go func(conn *grpc.ClientConn, id int) {
				numMsgs := *NUM_MSGS_PER_SAMPLE
				if *VARY_NUM_MSGS {
					numMsgs = rand.Int() % (*NUM_MSGS_PER_SAMPLE)
				}
				defer wg.Done()

				co := pb.NewEngageClient(conn)

				// Contact the server and print out its response.
				ctx, cancel := context.WithTimeout(context.Background(),
					time.Duration(*STREAM_CTX_TIMEOUT)*time.Second)
				defer cancel()

				start := time.Now()
				switch *OPTION {
				case "greet":
					r, err := co.Greet(ctx, &pb.Greeting{Name: c.name, Msg: "hi"})
					if err != nil || r.Msg != "hello "+c.name {
						log.Fatal(err)
					}

				case "shipdata":
					stream, err := co.ShipData(ctx,
						&pb.Request{Name: c.name, Ask: int32(numMsgs)})
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

				case "shipbulkdata":
					stream, err := co.ShipBulkData(ctx,
						&pb.Request{Name: c.name, Ask: int32(numMsgs)})
					if err != nil {
						log.Fatalf("[%v] could not request data to be bulk shipped: %v", c.name, err)
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

				default:
					log.Fatalf("Unknown OPTION: %v", *OPTION)
				}

				// Tachymeter is thread-safe.
				c.tachs[id].AddTime(time.Since(start))

			}(clientConn, k)
		}
	}
	wg.Wait()
}

func (c *client) terminate() {
	for _, conn := range c.conns {
		conn.Close()
	}
}

func (c *client) resultsStr() string {
	res := fmt.Sprintf("=======================================================\n"+
		"Client: %v\n"+
		"Addresses: %v\n"+
		"=======================================================\n",
		c.name, c.addresses)

	for _, tach := range c.tachs {
		res += tach.Calc().String() +
			"\n=======================================================\n"
	}

	return res
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
	fmt.Fprintf(file, "Connection concurrency: "+strconv.Itoa(*CONN_CONCURRENCY)+"\n")
	fmt.Fprintf(file, "Message type: "+(*OPTION)+"\n")
	if *OPTION != "greet" {
		fmt.Fprintf(file, "Num messages streamed per sample: "+
			strconv.Itoa(*NUM_MSGS_PER_SAMPLE)+"\n")
		fmt.Fprintf(file, "Varying number of messages per sample: "+
			strconv.FormatBool(*VARY_NUM_MSGS)+"\n")
	}
	fmt.Fprintf(file, "\n")
	for _, client := range clients {
		fmt.Fprintf(file, client.resultsStr()+"\n")
	}
	return nil
}

func main() {
	if *CPU_PROFILE != "" {
		f, err := os.Create(*CPU_PROFILE)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

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
