package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	"github.com/tckz/healthcheck"
	"github.com/tckz/healthcheck/api"
	"goji.io"
	"goji.io/pat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var myName string
var logger *logrus.Entry

func init() {
	myName = path.Base(os.Args[0])

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

type helloServer struct {
}

func (s *helloServer) SayHello(ctx context.Context, req *api.HelloRequest) (*api.HelloReply, error) {

	if rand.Intn(100) < 10 {
		panic("Random panic!!")
	}

	delay := rand.Intn(300)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	h, _ := os.Hostname()
	res := &api.HelloReply{
		Message: fmt.Sprintf("Hello %s, from %s", req.Name, h),
		Now:     TimestampPB(time.Now()),
	}
	return res, nil
}

func (s *helloServer) SayMorning(ctx context.Context, req *api.MorningRequest) (*api.MorningReply, error) {
	h, _ := os.Hostname()
	res := &api.MorningReply{
		Message: fmt.Sprintf("Morning %s, from %s", req.Name, h),
		Now:     TimestampPB(time.Now()),
	}
	return res, nil
}

func setupHealthCheckGateway(ctx context.Context, bindHealthCheck *string, conn *grpc.ClientConn) *http.Server {
	hcClient := healthpb.NewHealthClient(conn)

	logger := logger.WithFields(logrus.Fields{
		"type": "hc",
	})
	mux := goji.NewMux()
	mux.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(ow http.ResponseWriter, r *http.Request) {
			begin := time.Now()
			w := healthcheck.NewResponseWriterWrapper(ow)
			defer func() {
				logger := logger
				if r := recover(); r != nil {
					logger = logger.WithFields(logrus.Fields{
						"stack":         string(debug.Stack()),
						logrus.ErrorKey: r,
					})
					w.WriteHeader(http.StatusInternalServerError)
				}

				dur := time.Since(begin)
				ms := float64(dur) / float64(time.Millisecond)
				logger.WithFields(logrus.Fields{
					"status": w.StatusCode,
					"method": r.Method,
					"uri":    r.RequestURI,
					"remote": r.RemoteAddr,
					"msec":   ms,
				}).
					Infof("done: %s", dur)
			}()

			h.ServeHTTP(w, r)
		})
	})

	mux.HandleFunc(pat.Get("/healthz"), func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ctx.Done():
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			res, err := hcClient.Check(ctx, &healthpb.HealthCheckRequest{Service: "api.Greeter"})
			_ = res
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				logger.Infof("*** Failed to Check: %v", err)
			}
			fmt.Fprintf(w, "!\n")
		}
	})
	server := &http.Server{
		Addr:    *bindHealthCheck,
		Handler: mux,
	}

	return server

}

func main() {
	rand.Seed(time.Now().UnixNano())

	maxConnectionAge := flag.Duration("max-connection-age", 3600*time.Second, "Duration about gRPC connection refresh")
	delay := flag.Duration("delay", 30*time.Second, "Wait duration before shutdown")
	bind := flag.String("bind", ":3000", "addr:port")
	bindHealthCheck := flag.String("health-check", ":3001", "addr:port")
	healthCheckAddr := flag.String("health-check-server", "127.0.0.1:3000", "addr:port")
	flag.Parse()

	lis, err := net.Listen("tcp", *bind)
	if err != nil {
		logger.Fatalf("*** Failed to Listen(): %v", err)
	}

	grpc_logrus.ReplaceGrpcLogger(logger)

	gs := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge: *maxConnectionAge,
			Time:             150 * time.Second,
		}),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logger,
				grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
					return "grpc.time_ns", duration.Nanoseconds()
				})),
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
	)
	api.RegisterGreeterServer(gs, &helloServer{})
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(gs, healthServer)
	reflection.Register(gs)

	logger.Infof("Start to Serve: %s, %+v", lis.Addr(), gs.GetServiceInfo())
	go func() {
		if err := gs.Serve(lis); err != nil {
			logger.Fatalf("*** Failed to Serve(): %v", err)
		}
	}()
	healthServer.SetServingStatus("api.Greeter", healthpb.HealthCheckResponse_SERVING)

	conn, err := grpc.Dial(*healthCheckAddr,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithInsecure())
	if err != nil {
		logger.Fatalf("*** Failed to Dial %s: %v", *healthCheckAddr, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	server := setupHealthCheckGateway(ctx, bindHealthCheck, conn)
	if err != nil {
		logger.Fatalf("*** Failed to setupHealthCheckGateway: %v", err)
	}
	defer cancel()

	logger.Infof("Start to serve HealthCheck gateway: %s", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("*** Failed to ListenAndServe(): %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigCh
	logger.Infof("Receive signal: %v, Wait %s before shutdown", sig, *delay)
	cancel()
	time.Sleep(*delay)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	gs.GracefulStop()

	logger.Info("exit")
}
