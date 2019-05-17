module github.com/tckz/healthcheck

go 1.12

replace github.com/tckz/vegetahelper => ../vegetahelper

require (
	github.com/goji/glogrus v0.0.0-20171018100434-f7c99b3e8e6f
	github.com/golang/protobuf v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/sirupsen/logrus v1.4.1
	github.com/tckz/vegetahelper v0.0.2
	github.com/tsenart/vegeta v12.3.0+incompatible
	github.com/zenazn/goji v0.9.1-0.20160507202103-64eb34159fe5
	goji.io v2.0.2+incompatible
	golang.org/x/net v0.0.0-20190509222800-a4d6f7feada5
	google.golang.org/grpc v1.20.1
)
