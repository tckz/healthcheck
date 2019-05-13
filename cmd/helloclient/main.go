package main

import (
	"flag"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"github.com/tckz/healthcheck/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	defaultName = "world"
)

func main() {
	timeOutSec := flag.Int("timeout", 3, "Seconds to timeout")
	retry := flag.Uint("retry", 3, "Max retry")
	server := flag.String("server", "127.0.0.1:3000", "Server addr:port")
	flag.Parse()

	logrus.SetFormatter(&logrus.JSONFormatter{
		// GKE向けのフィールド名に置換え
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
	})
	logrus.Infof("Server: %s", *server)

	conn, err := grpc.Dial(*server,
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(*retry),
		)),
		grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("*** Failed to Dial %s: %v", *server, err)
	}
	defer conn.Close()
	client := api.NewGreeterClient(conn)

	name := defaultName
	if flag.NArg() >= 1 {
		name = flag.Arg(0)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeOutSec)*time.Second)
	defer cancel()
	r, err := client.SayHello(ctx, &api.HelloRequest{Name: name})
	if err != nil {
		logrus.Fatalf("*** Failed to SayHello: %v", err)
	}

	now := time.Unix(r.Now.Seconds, int64(r.Now.Nanos))
	logrus.Printf("Response: Message=%s, Now=%s", r.Message, now)
}
