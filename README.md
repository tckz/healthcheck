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

* Place protoc command and includes, Install protoc-gen-go, Generate .pb.go

    ```bash
    $ make gen
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

## healthy-grpc

* gRPC server that serving api.Greeter
* api.Greeter.SayHello() randomly raise panic.

## helloclient

* gRPC client for invoke api.Greeter.SayHello()

## helloattacker

* Load test client about api.Greeter.SayHello
* It uses vegeta as library.


# My environment

* CentOS Stream 8 x64
* GNU Make 4.2.1
* Go 1.20.6
* Docker CE 24.0.5
* Docker Compose 2.20.2
