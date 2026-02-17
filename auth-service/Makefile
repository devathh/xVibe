make keys:
	openssl genrsa -out ./jwtRS256.key 2048
	openssl rsa -in ./jwtRS256.key -pubout -out ./jwtRS256.key.pub

make mtls:
	mkdir -p ./certs ./client-certs
	mkcert -install
	mkcert auth-service 127.0.0.1 localhost

	mv auth-service+2.pem ./certs/server.crt
	mv auth-service+2-key.pem ./certs/server.key

	mkcert -client "auth-gateway"
	mv auth-gateway-client.pem ./client-certs/client.crt
	mv auth-gateway-client-key.pem ./client-certs/client.key

	@CA_DIR=$$(mkcert -CAROOT | tr -d '\r\n'); \
	cp "$$CA_DIR/rootCA.pem" ./certs/
	cp ./certs/rootCA.pem ./client-certs/