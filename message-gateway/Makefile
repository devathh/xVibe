mkcertlocal:
	mkcert -install
	mkcert localhost 127.0.0.1
	
	mkdir ./certs/
	mv localhost+1.pem server.crt
	mv ./server.crt ./certs/
	mv localhost+1-key.pem server.key 
	mv ./server.key ./certs/
