.PHONY: init
init:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: bootstrap
bootstrap:
	cd ./deploy/docker-compose && docker compose up -d && cd ../../
	go run ./cmd/migration
	nunu run ./cmd/server

.PHONY: mock
mock:
	mockgen -source=internal/service/user.go -destination test/mocks/service/user.go
	mockgen -source=internal/repository/user.go -destination test/mocks/repository/user.go
	mockgen -source=internal/repository/repository.go -destination test/mocks/repository/repository.go


.PHONY: produce
produce:
	echo "produce to server ..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server ./cmd/server
	ssh ubuntu@3.15.54.60 "sudo supervisorctl stop server"
	scp bin/server ubuntu@3.15.54.60:/home/ubuntu/SOL/api
	ssh ubuntu@3.15.54.60 "sudo supervisorctl start server"


.PHONY: test
test:
	echo "produce to server ..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server ./cmd/server
	ssh ubuntu@3.18.34.28 "sudo supervisorctl stop server"
	scp bin/server ubuntu@3.18.34.28:/home/ubuntu/SOL/api
	ssh ubuntu@3.18.34.28 "sudo supervisorctl start server"


.PHONY: build
build:
	go build -ldflags="-s -w" -o ./server ./cmd/server

.PHONY: docker
docker:
	docker build -f deploy/build/Dockerfile --build-arg APP_RELATIVE_PATH=./cmd/task -t 1.1.1.1:5000/demo-task:v1 .
	docker run --rm -i 1.1.1.1:5000/demo-task:v1

.PHONY: swag
swag:
	swag init  -g cmd/server/main.go -o ./docs --parseDependency
