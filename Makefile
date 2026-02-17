PROTO_SERVICE=?
PROTO_CLIENT=?
PROTO_VERSION=v1

ENV_DIRS = ./auth-service ./auth-gateway ./chat-service ./chat-gateway ./message-service ./message-gateway ./rate-limiter
SOURCE_ENV = dot.env

# generate service proto
proto:
	protoc --proto_path=./contracts/ \
		--go_out=./$(PROTO_SERVICE)-service/api --go_opt=paths=source_relative \
		--go-grpc_out=./$(PROTO_SERVICE)-service/api --go-grpc_opt=paths=source_relative \
		./contracts/$(PROTO_SERVICE)/$(PROTO_VERSION)/$(PROTO_SERVICE).proto

protoclient:
	protoc --proto_path=./contracts/ \
		--go_out=./$(PROTO_CLIENT)/api --go_opt=paths=source_relative \
		--go-grpc_out=./$(PROTO_CLIENT)/api --go-grpc_opt=paths=source_relative \
		./contracts/$(PROTO_SERVICE)/$(PROTO_VERSION)/$(PROTO_SERVICE).proto


quickenv:
	@for dir in $(ENV_DIRS); do \
		if [ -d "$$dir" ]; then \
			echo "handling $$dir..."; \
			cp $$dir/$(SOURCE_ENV) "$$dir/.env"; \
		fi \
	done

quickstart:
	make quickenv
	docker-compose up -d