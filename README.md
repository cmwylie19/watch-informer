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
    
    
```bash
mockgen -destination=mocks/mock_dynamic.go -package=mocks k8s.io/client-go/dynamic Interface
mockgen -destination=mocks/mock_watchevents.go -package=mocks watch-informer/api Watcher_WatchEventsServer
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
grpcurl -plaintext localhost:50051 list api.Watcher

# Describe a method
grpcurl -plaintext localhost:50051 describe api.Watcher.[StartWatch/WatchEvents]

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
localhost:50051 api.Watcher.StartWatch

# Start the watch 
grpcurl -plaintext -d '{"session_id": "-v1-pods"}' \
localhost:50051 api.Watcher.WatchEvents
```

