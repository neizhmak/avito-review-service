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

test:
	go test -v ./...

lint:
	golangci-lint run

load-test:
	docker run --rm -i \
		-v $(PWD)/k6-script.js:/script.js \
		--add-host=host.docker.internal:host-gateway \
		grafana/k6 run /script.js