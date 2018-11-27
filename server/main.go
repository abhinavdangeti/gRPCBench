package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/abhinavdangeti/gRPCBench/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var MAX_CONCURRENT_STREAMS = flag.Int("maxConcurrentStreams", 0, "int")

var greetings map[string]string

func init() {
	flag.Parse()

	// initialize greetings
	greetings = make(map[string]string)
	greetings["hi"] = "hello"
	greetings["bye"] = "goodbye"
}

var sample_json_data = string(`{
	"status":{
		"total":1,
		"failed":0,
		"successful":1,
		"errors":{}
	},
	"request":{
		"query":{
			"query":"blah blah blah"
		},
		"size":10,
		"from":0,
		"highlight":null,
		"fields":[],
		"facets":null,
		"explain":false
	},
	"hits":[
		{
			"id": "doc_1"
		},
		{
			"id": "doc_2"
		},
		{
			"id": "doc_3"
		}
	],
	"total_hits":3,
	"max_score":0,
	"took":0,
	"facets":null
}`)

type server struct {
	name string
}

func (s *server) Greet(ctx context.Context, in *pb.Greeting) (*pb.Greeting, error) {
	if in == nil {
		return nil, fmt.Errorf("-no greeting-")
	}

	reply := "didn't recoginize that,"
	if v, exists := greetings[in.Msg]; exists {
		reply = v
	}

	return &pb.Greeting{Name: s.name, Msg: reply + " " + in.Name}, nil
}

func (s *server) IdentifyData(stream pb.Engage_IdentifyDataServer) error {

	return nil
}

func (s *server) ShipData(req *pb.Request, stream pb.Engage_ShipDataServer) error {
	if req == nil {
		return fmt.Errorf("-no request-")
	}
	for i := 0; i < int(req.Ask); i++ {
		resp := &pb.Response{Name: s.name, Content: sample_json_data}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) ShipBulkData(req *pb.Request, stream pb.Engage_ShipBulkDataServer) error {
	if req == nil {
		return fmt.Errorf("-no request-")
	}
	responses := make([]*pb.Response, int(req.Ask))
	for i := 0; i < int(req.Ask); i++ {
		responses[i] = &pb.Response{Name: s.name, Content: sample_json_data}
	}
	return stream.Send(&pb.Responses{Items: responses})
}

func main() {
	var wg sync.WaitGroup
	for _, val := range []string{":12345", ":23456", ":34567"} {
		wg.Add(1)
		go func(port string) {
			defer wg.Done()
			listen, err := net.Listen("tcp", port)
			if err != nil {
				log.Fatalf("Failed to listen: %v", err)
			}
			s := grpc.NewServer()
			if *MAX_CONCURRENT_STREAMS > 0 {
				s = grpc.NewServer(grpc.MaxConcurrentStreams(uint32(*MAX_CONCURRENT_STREAMS)))
			}
			pb.RegisterEngageServer(s, &server{name: "server" + port})
			// register reflection service on gRPC server.
			reflection.Register(s)
			log.Printf("Serving %v", port)
			if err = s.Serve(listen); err != nil {
				log.Fatalf("Failed to serve: %v", err)
			}
		}(val)
	}
	wg.Wait()
}
