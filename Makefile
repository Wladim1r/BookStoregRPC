gen:
	protoc --proto_path=proto --go_out=pb --go-grpc_out=pb proto/*.proto

docker-up:
	docker compose up

docker-down:
	docker compose down -v

client:
	go run cmd/client/client.go