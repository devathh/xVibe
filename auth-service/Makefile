make keys:
	openssl genrsa -out ./jwtRS256.key 2048
	openssl rsa -in ./jwtRS256.key -pubout -out ./jwtRS256.key.pub

make mtls:
	mkcert -install
	mkcert xvibe-auth-service 127.0.0.1
	
	mkdir ./certs/
	mv xvibe-auth-service+1.pem server.crt
	mv ./server.crt ./certs/
	mv xvibe-auth-service+1-key.pem server.key 
	mv ./server.key ./certs/

	mkdir ./client-certs
	mkcert -client "xvibe-auth-gateway"
	mv xvibe-auth-gateway-client.pem client.crt
	mv client.crt ./client-certs/
	mv xvibe-auth-gateway-client-key.pem client.key
	mv client.key ./client-certs/

	@CA_DIR=$$(mkcert -CAROOT | tr -d '\r\n'); \
	cp "$$CA_DIR/rootCA.pem" ./certs/
	cp ./certs/rootCA.pem ./client-certs/