# Watch Informer

- [Watch Informer](#watch-informer)
  - [Usage](#usage)
  - [Test](#test)
  - [Generic Usage](#generic-usage)
  - [Generate the Protocol Buffers](#generate-the-protocol-buffers)
  - [Generate Mocks](#generate-mocks)

A simple gRPC server that watches for Kubernetes resources and streams events to clients. It can be run in or out of cluster (for pepr dev).

```bash
Starts the watch-informer gRPC server

Usage:
  watch-informer [flags]

Flags:
  -h, --help               help for watch-informer
      --in-cluster         Use in-cluster configuration (default true)
  -l, --log-level string   Log level (debug, info, error) (default "info")
```



## Usage

Bring up a dev cluster with application deployed  
```bash
make deploy-dev
```

Get Events

```bash
make curl-dev
```


## Test 

unit 
```bash
make unit test
```

e2e
```bash
make e2e test
```

## Generic Usage  

Server  

```bash
go run main.go --in-cluster=false
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

# Start the watch 
grpcurl -plaintext -d '{"group": "", "version": "v1", "resource": "pod", "namespace": "default"}' \
localhost:50051 api.WatchService.Watch

# Start the watch in cluster
kubectl exec -it curler -- grpcurl -plaintext -d '{"group": "", "version": "v1", "resource": "pod", "namespace": "default"}' watch-informer.watch-informer.svc.cluster.local:50051 api.WatchService.Watch
```

## Generate the ProtoBufs

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/apiv1.proto
```

TypeScript Client Proto files

```bash
npx grpc_tools_node_protoc \
  --proto_path=src/api \
  --plugin=protoc-gen-ts=node_modules/.bin/protoc-gen-ts \
  --plugin=protoc-gen-grpc=node_modules/.bin/grpc_tools_node_protoc_plugin \
  --js_out=import_style=commonjs,binary:src/api \
  --ts_out=src/api \
  --grpc_out=grpc_js:src/api \
  src/api/apiv1.proto
```

## Generate Mocks
// go list -m -f '{{.Dir}}' k8s.io/client-go
```bash
mockgen -source=/Users/cmwylie19/go/pkg/mod/k8s.io/client-go@v0.31.0/dynamic/interface.go -destination=mocks/mock_dynamic.go -package=mocks k8s.io/client-go/dynamic Interface
mockgen -source=pkg/server/server.go -destination=mocks/mock_watch_service.go -package=mocks github.com/cmwylie19/watch-informer/api WatchService_WatchServer

mockgen -destination=mocks/mock_api.go -package=mocks github.com/cmwylie19/watch-informer/api WatchService_WatchServer
mockgen -destination mocks/mock_logging.go -package mocks -source ./pkg/logging/logging.go
mockgen -source=./api/apiv1.pb.go -destination=./mocks/apiv1.pb.go -package=mocks
```
