package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"github.com/tckz/healthcheck/api"
	vh "github.com/tckz/vegetahelper"
	vhgrpc "github.com/tckz/vegetahelper/grpc"
	vegeta "github.com/tsenart/vegeta/lib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var myName string
var logger *logrus.Entry

func init() {
	myName = filepath.Base(os.Args[0])

	logrus.SetFormatter(&logrus.JSONFormatter{
		// GKE向けのフィールド名に置換え
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
	})

	logger = logrus.WithFields(logrus.Fields{
		"app": myName,
	})

}

func openOutFile(out string) (*os.File, func()) {
	switch out {
	case "stdout":
		return os.Stdout, func() {}
	default:
		if f, err := os.Create(out); err != nil {
			logger.Fatalf("*** Failed to Open: %v", err)
			return nil, func() {}
		} else {
			return f, func() { f.Close() }
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	rate := vegeta.Rate{
		Freq: 10,
		Per:  1 * time.Second,
	}
	duration := flag.Duration("duration", 10*time.Second, "Duration of the test [0 = forever]")
	flag.Var(&vh.RateFlag{&rate}, "rate", "Number of requests per time unit")
	output := flag.String("output", "stdout", "Output file")
	workers := flag.Uint64("workers", vegeta.DefaultWorkers, "Initial number of workers")

	keepAlivePeriod := flag.Duration("keepalive-period", 150*time.Second, "Keepalive period of gRPC connection")
	server := flag.String("server", "127.0.0.1:3000", "Server addr:port")
	retry := flag.Uint("retry", 3, "Max retry")
	flag.Parse()

	logger.Infof("Server: %s", *server)

	conn, err := grpc.Dial(*server,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                *keepAlivePeriod,
			PermitWithoutStream: true,
		}),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(*retry),
		)),
		grpc.WithStatsHandler(&vhgrpc.RpcStatsHandler{}))
	if err != nil {
		logger.Fatalf("*** Failed to Dial %s: %v", *server, err)
	}
	defer conn.Close()
	client := api.NewGreeterClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	atk := vh.NewAttacker(
		func(ctx context.Context) (*vh.HitResult, error) {
			return vhgrpc.HitGrpc(ctx, func(ctx context.Context) error {
				_, err := client.SayHello(ctx, &api.HelloRequest{
					Name: "oreore",
				})
				return err
			})
		},
		vh.WithWorkers(*workers))
	res := atk.Attack(ctx, rate, *duration, "hello")
	out, closer := openOutFile(*output)
	defer closer()
	enc := vegeta.NewEncoder(out)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

loop:
	for {
		select {
		case s := <-sig:
			logger.Infof("Received signal: %s", s)
			cancel()
			break loop
		case r, ok := <-res:
			if !ok {
				break loop
			}
			if err := enc.Encode(r); err != nil {
				logger.Errorf("*** Failed to Encode: %v", err)
				break loop
			}
		}
	}
}
