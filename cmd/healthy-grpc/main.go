package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/tckz/healthcheck"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/tckz/healthcheck/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
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

type helloServer struct {
}

func (s *helloServer) SayHello(ctx context.Context, req *api.HelloRequest) (*api.HelloReply, error) {
	res := &api.HelloReply{
		Message: fmt.Sprintf("Hello %s, from %s", req.Name, os.Getenv("HOSTNAME")),
		Now:     healthcheck.TimestampPB(time.Now()),
	}
	return res, nil
}

func (s *helloServer) SayMorning(ctx context.Context, req *api.MorningRequest) (*api.MorningReply, error) {
	res := &api.MorningReply{
		Message: fmt.Sprintf("Morning %s, from %s", req.Name, os.Getenv("HOSTNAME")),
		Now:     healthcheck.TimestampPB(time.Now()),
	}
	return res, nil
}

func main() {
	bind := flag.String("bind", ":3000", "addr:port")
	flag.Parse()

	lis, err := net.Listen("tcp", *bind)
	if err != nil {
		logger.Fatalf("*** Failed to Listen(): %v", err)
	}

	grpc_logrus.ReplaceGrpcLogger(logger)

	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ns", duration.Nanoseconds()
		}),
	}

	gs := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logger, logrusOpts...),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				defer func() {
					if r := recover(); r != nil {
						err = status.Errorf(codes.Internal, "%v", r)
						ctxlogrus.AddFields(ctx, logrus.Fields{
							"stack": string(debug.Stack()),
						})
					}
				}()

				return handler(ctx, req)
			},
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_logrus.StreamServerInterceptor(logger, logrusOpts...),
			func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				defer func() {
					if r := recover(); r != nil {
						err = status.Errorf(codes.Internal, "%v", r)
						ctxlogrus.AddFields(ss.Context(), logrus.Fields{
							"stack": string(debug.Stack()),
						})
					}
				}()

				return handler(srv, ss)
			},
		),
	)
	api.RegisterGreeterServer(gs, &helloServer{})
	reflection.Register(gs)

	logger.Infof("Start to Serve: %s", lis.Addr())
	go func() {
		if err := gs.Serve(lis); err != nil {
			logger.Fatalf("*** Failed to Serve(): %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigCh
	logger.Infof("Receive signal: %v", sig)
	gs.GracefulStop()

}
