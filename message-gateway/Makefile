mkcertlocal:
	mkcert -install
	mkcert xvibe-message-gateway 127.0.0.1
	
	mkdir ./certs/
	mv xvibe-message-gateway+1.pem server.crt
	mv ./server.crt ./certs/
	mv xvibe-message-gateway+1-key.pem server.key 
	mv ./server.key ./certs/
