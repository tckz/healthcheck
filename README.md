healthcheck
===

# Requirements

* Go 1.20
* GNU Make
* Docker
* Docker Compose
* protoc  
  https://github.com/google/protobuf

# Before build

* Place protoc command and includes  
  https://github.com/google/protobuf/releases
  
    ```
    usvc/
      README.md
      protoc/  <-- Extract the archive into this dir.
        bin/
          protoc
        include/
          google/protobuf/*.proto
    ```
* Install protoc-gen-go

    ```bash
    $ go get -u github.com/golang/protobuf/protoc-gen-go
    ```

# Build

* Build modules
```bash
# build
$ make
```

* Build docker images
```bash
$ make images
```

# Run

* Run servers.
```bash
$ docker-compose up
```

* e.g.) Run load test.
```bash
$ go run ./cmd/helloattacker/ -server localhost:3002 -duration 3m -rate 100/s > results.bin
$ vegeta report < results.bin
```

# Modules

## healthy-old-goji

* Web server that using goji which is older version.

## healthy-grpc

* gRPC server that serving api.Greeter
* api.Greeter.SayHello() randomly raise panic.

## helloclient

* gRPC client for invoke api.Greeter.SayHello()

## helloattacker

* Load test client about api.Greeter.SayHello
* It uses vegeta as library.


# My environment

* CentOS 7 x64
* GNU Make 3.82
* Go 1.12.5
* Docker CE 18.09.3
* Docker Compose 1.23.1
