make mtls:
	mkcert -install
	mkcert localhost 127.0.0.1
	
	mkdir ./certs/
	mv localhost+1.pem server.crt
	mv ./server.crt ./certs/
	mv localhost+1-key.pem server.key 
	mv ./server.key ./certs/

	mkcert -client "xvibe-chat-gateway"
	mv xvibe-chat-gateway-client.pem client.crt
	mv xvibe-chat-gateway-client-key.pem client.key

	@CA_DIR=$$(mkcert -CAROOT | tr -d '\r\n'); \
	cp "$$CA_DIR/rootCA.pem" ./certs/