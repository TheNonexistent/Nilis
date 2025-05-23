PROTO_DIR = pkg/store
PACKAGE = $(shell head -1 go.mod | awk '{print $$2}')

run: build
	@./bin/nilis

build:
	go build -o bin/nilis ./cmd/server

generate:
	protoc -I${PROTO_DIR} --go_opt=module=${PACKAGE} --go_out=. ${PROTO_DIR}/*.proto --go-grpc_opt=module=${PACKAGE} --go-grpc_out=. ${PROTO_DIR}/*.proto
