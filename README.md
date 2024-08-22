# Watch Informer

## Prereqs

Go Plugins for Protocol Buffers  

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

## Generate the Protocol Buffers

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/apiv1.proto
```

## Generate Mocks
    
go list -f '{{.Dir}}' k8s.io/client-go/dynamic
```bash

mockgen -destination=mocks/mock_api.go -package=mocks github.com/cmwylie19/watch-informer/api WatchService_WatchServer

mockgen -destination=mocks/mock_dynamic.go -package=mocks -source $GOPATH/pkg/mod/k8s.io/client-go@v0.31.0/dynamic/interface.go
mockgen -destination mocks/mock_logging.go -package mocks -source ./pkg/logging/logging.go

mockgen -source=./api/apiv1.pb.go -destination=./mocks/apiv1.pb.go -package=mocks
```

## Test 

```bash
go test ./...  
```

## Generic Usage  

Server  

```bash
go run main.go --group "" --version "v1" --resource "pods" --namespace "default"

go run main.go
```


Client

```bash
# List services
grpcurl -plaintext localhost:50051 list

# List methods in a service
grpcurl -plaintext localhost:50051 list api.WatchService

# Describe a method
grpcurl -plaintext localhost:50051 describe api.WatchService.Watch

# Invoke a method
grpcurl -plaintext -d '{"group": "", "version": "v1", "resource": "pods", "namespace": "default"}' \
    localhost:50051 <service-name>.<method-name>

```


## Usage 

Server

```bash
go run main.go
```

Client

```bash
# Configure the server
grpcurl -plaintext -d '{"group": "", "version": "v1", "resource": "pods", "namespace": "default"}' \
localhost:50051 api.WatchService.Watch

# Start the watch 
grpcurl -plaintext -d '{"session_id": "-v1-pods"}' \
localhost:50051 api.Watcher.WatchEvents
```

