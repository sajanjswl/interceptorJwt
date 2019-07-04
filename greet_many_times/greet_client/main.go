package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/DnsProject/greet_many_times/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/testdata"
)

var addr = flag.String("addr", "localhost:50051", "the address to connect to")

//jwt interceptor
type (
	JWTInterceptor struct {
		http     *http.Client // The HTTP client for calling the token-serving API
		token    string       // The JWT token that will be used in every call to the server
		username string       // The username for basic authentication
		password string       // The password for basic authentication
		endpoint string       // The HTTP endpoint to hit to obtain tokens
	}
)

func (jwt *JWTInterceptor) streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

	authCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "bearer "+jwt.token)

	fmt.Println("Hello from ClientStreamInterceptor")

	s, err := streamer(authCtx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func main() {

	fmt.Println("Hello from client")
	flag.Parse()

	jwt := &JWTInterceptor{
		// Set up all the members here
		token: "pass",
	}

	creds, err := credentials.NewClientTLSFromFile(testdata.Path("ca.pem"), "x.test.youtube.com")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	cc, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds), grpc.WithStreamInterceptor(jwt.streamInterceptor))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	doGreetManyTimes(c)
}

func doUnary(c greetpb.GreetServiceClient) {

	fmt.Println("starting to do Unary RPC")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Veer",
			LastName:  "Jaiswal",
		},
	}
	res, err := c.Greet(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling greet Rpc %v", err)
	}
	log.Printf("Response from greet : %v", res.Result)
}

func doGreetManyTimes(c greetpb.GreetServiceClient) {

	fmt.Println("starting to do a server Streamin  RPC...")

	fmt.Println("Hello from client Greet Stream Call")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Veer",
			LastName:  "Jaiswal",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling greet Rpc %v", err)
	}
	for {
		msg, err := resStream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream %v", err)
		}
		log.Printf("Response from greetManyTimes : %v", msg.Result)
	}

}
