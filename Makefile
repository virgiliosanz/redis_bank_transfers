init:
	@go run cmd/init/main.go

broker:
	@go run cmd/broker/main.go

consume:
	@go run cmd/consume/main.go

build:
	@go build -v -o bin/init cmd/init/main.go
	@go build -v -o bin/broker cmd/broker/main.go
	@go build -v -o bin/consume cmd/consume/main.go

clean:
	@rm -f bin/*
