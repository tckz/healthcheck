package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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

func main() {
	healthCheckAddr := flag.String("health-check-addr", "127.0.0.1:3000", "addr:port")
	service := flag.String("service", "api.Greeter", "Service name which is checked healthy or not")
	optInsecure := flag.Bool("insecure", false, "Use http instead of https")
	flag.Parse()

	var grpcOpts []grpc.DialOption
	if *optInsecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	conn, err := grpc.Dial(*healthCheckAddr, grpcOpts...)
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
