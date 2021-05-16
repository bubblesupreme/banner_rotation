start:
	docker-compose build
	docker-compose --env-file=.test-env up -d
	bash wait_service_up.sh

stop:
	docker-compose down

lint:
	golangci-lint run

test: stop start
	go test -v -count=1 -race -timeout=1m ./...
	docker-compose down