test:
	go test -v -count=1 -race -timeout=1m ./...

start:
	docker-compose build
	docker-compose --env-file=.test-env up

stop:
	docker-compose down

lint:
	golangci-lint run
