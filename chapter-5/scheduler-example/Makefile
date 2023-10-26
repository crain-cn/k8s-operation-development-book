
.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: build
build:  fmt vet ## Build kube-scheduler binary.
	go build -o bin/kube-scheduler main.go

.PHONY: clean
clean:
	rm -rf bin/*
