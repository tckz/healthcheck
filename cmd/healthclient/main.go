package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

var myName string
var logger *logrus.Entry

func init() {
	myName = path.Base(os.Args[0])

	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
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

func main() {
	healthCheckAddr := flag.String("health-check-addr", "127.0.0.1:3000", "addr:port")
	service := flag.String("service", "api.Greeter", "Service name to health check")
	flag.Parse()

	conn, err := grpc.Dial(*healthCheckAddr,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             2 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithInsecure())
	if err != nil {
		logger.Fatalf("*** Failed to Dial %s: %v", *healthCheckAddr, err)
	}
	defer conn.Close()
	hcClient := healthpb.NewHealthClient(conn)
	ctx := context.Background()
	res, err := hcClient.Check(ctx, &healthpb.HealthCheckRequest{Service: *service})
	if err != nil {
		logger.Fatalf("*** Failed to Check: %v", err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", res)
}
