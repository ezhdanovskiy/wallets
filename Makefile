APP_NAME=wallets
CUR_DIR=$(shell pwd)
SRC=$(CUR_DIR)/cmd
BINARY_NAME=$(CUR_DIR)/bin/$(APP_NAME)

.PHONY: generate fmt test test/int build run clean mod/tidy build-container run-container

all: fmt generate test build clean mod/tidy

generate:
	$(info ************ GENERATE MOCKS ************)
	go generate -v ./...

fmt:
	$(info ************ RUN FROMATING ************)
	go fmt ./...

test:
	$(info ************ RUN UNIT TESTS ************)
	go test -count=1 -v ./...

test/int:
	$(info ************ RUN UNIT AND INTEGRATION TESTS ************)
	go test -count=1 -tags integration -v ./...

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

mod/tidy:
	$(info ************ MOD TIDY ************)
	go mod tidy

build/docker:
	$(info ************ BUILD CONTAINER ************)
	docker build -t $(APP_NAME) .
run/docker:
	$(info ************ RUN CONTAINER ************)
	docker run --rm --env DB_HOST=host.docker.internal -p 8080:8080 --name $(APP_NAME) $(APP_NAME)

migrate/up:
	$(info ************ MIGRATE UP ************)
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable" -verbose up
migrate/down:
	$(info ************ MIGRATE DOWN ************)
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable" -verbose down 2

postgres/up:
	$(info ************ UP POSTGRES IN DOCKER-COMPOSE ************)
	docker-compose up -d postgres
postgres/down:
	$(info ************ DOWN DOCKER-COMPOSE ************)
	docker-compose down --remove-orphans

test/int/docker-compose: postgres/up test/int postgres/down
