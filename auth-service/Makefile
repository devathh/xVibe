make keys:
	openssl genrsa -out ./jwtRS256.key 2048
	openssl rsa -in ./jwtRS256.key -pubout -out ./jwtRS256.key.pub

make mtls:
	mkcert -install
	mkcert localhost 127.0.0.1
	
	mkdir ./certs/
	mv localhost+1.pem server.crt
	mv ./server.crt ./certs/
	mv localhost+1-key.pem server.key 
	mv ./server.key ./certs/

	mkcert -client "xvibe-auth-gateway"
	mv xvibe-auth-gateway-client.pem client.crt
	mv xvibe-auth-gateway-client-key.pem client.key

	@CA_DIR=$$(mkcert -CAROOT | tr -d '\r\n'); \
	cp "$$CA_DIR/rootCA.pem" ./certs/