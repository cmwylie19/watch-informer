.DEFAULT_GOAL := docker-image
CI_IMAGE ?= watch-informer:ci
IMAGE ?= watch-informer:dev
PROD_IMAGE ?= cmwylie19/watch-informer:prod

build-ci-image:
	docker build -t $(CI_IMAGE) -f Dockerfile .
build-push-prod-image:
	docker buildx build --platform linux/amd64,linux/arm64 --push -t $(PROD_IMAGE) -f Dockerfile.amd .

build-dev-image: 
	docker build -t $(IMAGE) -f Dockerfile .

build-prod-image:
	docker build -t $(IMAGE) -f Dockerfile.amd .

build-push-arm-image: 
	docker buildx build --push -t $(IMAGE) -f Dockerfile .

build-push-amd-image: 
	docker buildx build --push -t $(IMAGE) -f Dockerfile.amd .

unit-test:
	go test -v ./... -tags='!e2e'

e2e-test:
	ginkgo -v --tags='e2e' ./e2e

deploy-dev:
	kind create cluster
	docker build -t watch-informer:dev . -f Dockerfile
	kind load docker-image watch-informer:dev
	kubectl apply -k kustomize/overlays/dev

curl-dev:
	docker build -t curler:ci -f hack/Dockerfile hack/
	kind load docker-image curler:ci
	kubectl apply -f hack/
	kubectl wait --for=condition=ready pod -n watch-informer -l app=curler
	kubectl exec -it curler -n watch-informer -- grpcurl -plaintext -d '{"group": "", "version": "v1", "resource": "pod", "namespace": "watch-informer"}' watch-informer.watch-informer.svc.cluster.local:50051 api.WatchService.Watch | jq

clean-dev:
	kind delete cluster --name kind
	docker system prune -a -f

check-logs:
	kubectl logs -n watch-informer -l app=watch-informer -f | jq
