APP_NAME=wallets
CUR_DIR=$(shell pwd)
SRC=$(CUR_DIR)/cmd
BINARY_NAME=$(CUR_DIR)/bin/$(APP_NAME)

.PHONY: test build run clean build-container run-container

all: fmt generate test build

generate:
	$(info ************ GENERATE MOCKS ************)
	go generate -v ./...

fmt:
	$(info ************ RUN FROMATING ************)
	go fmt ./...

test:
	$(info ************ RUN TESTS ************)
	go test -v ./...

build:
	$(info ************ BUILD ************)
	CGO_ENABLED=0 go build -o $(BINARY_NAME) -v $(SRC)

run:
	$(info ************ CLEAN ************)
	$(BINARY_NAME)

clean:
	$(info ************ CLEAN ************)
	go clean
	rm -f $(BINARY_NAME)

build-container:
	$(info ************ BUILD CONTAINER ************)
	docker build -t $(APP_NAME) .

run-container:
	$(info ************ RUN CONTAINER ************)
	docker run --rm --env DB_HOST=host.docker.internal -p 8080:8080 --name $(APP_NAME) $(APP_NAME)
