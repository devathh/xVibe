make mtls:
	mkcert -install
	mkcert xvibe-message-service 127.0.0.1 localhost
	
	mkdir ./certs/
	mv xvibe-message-service+2.pem server.crt
	mv ./server.crt ./certs/
	mv xvibe-message-service+2-key.pem server.key 
	mv ./server.key ./certs/

	mkdir ./gateway-certs/
	mkcert -client "xvibe-chat-gateway"
	mv xvibe-chat-gateway-client.pem client.crt
	mv client.crt ./gateway-certs/
	mv xvibe-chat-gateway-client-key.pem client.key
	mv client.key ./gateway-certs/

	@CA_DIR=$$(mkcert -CAROOT | tr -d '\r\n'); \
	cp "$$CA_DIR/rootCA.pem" ./certs/
	cp ./certs/rootCA.pem ./gateway-certs/