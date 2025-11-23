TEST_DSN="postgres://user:password@localhost:5433/reviewer_db_test?sslmode=disable"

run:
	docker-compose up --build

build:
	docker-compose build

up:
	docker-compose up

db:
	docker-compose up -d db

stop:
	docker-compose down

clean:
	docker-compose down -v

deps:
	go install github.com/pressly/goose/v3/cmd/goose@latest

test:
	docker-compose up -d db_test
	sleep 5

	GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(TEST_DSN) goose -dir ./migrations up

	TEST_DB_CONNECTION_STRING=$(TEST_DSN) go test -v ./...

	docker-compose down -v db_test

lint:
	golangci-lint run

load-test:
	docker run --rm -i \
		-v $(PWD)/k6-script.js:/script.js \
		--add-host=host.docker.internal:host-gateway \
		grafana/k6 run /script.js