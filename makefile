RUN_PATH=cmd/api/main.go
DEV_PATH=cmd/run-dev/main.go

dev:
	go run ./$(DEV_PATH)

run:
	go run ./$(RUN_PATH)

see:
	docker ps