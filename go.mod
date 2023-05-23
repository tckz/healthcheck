module github.com/tckz/healthcheck

go 1.20

replace github.com/tckz/vegetahelper => ../vegetahelper

require (
	github.com/golang/protobuf v1.5.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/sirupsen/logrus v1.9.2
	github.com/tckz/vegetahelper v0.0.2
	github.com/tsenart/vegeta/v12 v12.8.4
	go.uber.org/zap v1.24.0
	goji.io v2.0.2+incompatible
	golang.org/x/net v0.10.0
	google.golang.org/grpc v1.55.0
)

require (
	github.com/influxdata/tdigest v0.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
