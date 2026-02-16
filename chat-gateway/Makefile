mkcertlocal:
	mkcert -install
	mkcert xvibe-chat-gateway 127.0.0.1
	
	mkdir ./certs/
	mv xvibe-chat-gateway+1.pem server.crt
	mv ./server.crt ./certs/
	mv xvibe-chat-gateway+1-key.pem server.key 
	mv ./server.key ./certs/
