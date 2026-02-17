mkcertlocal:
	mkcert -install
	mkcert auth-gateway 127.0.0.1 localhost
	
	mkdir ./certs/
	mv auth-gateway+2.pem server.crt
	mv ./server.crt ./certs/
	mv auth-gateway+2-key.pem server.key 
	mv ./server.key ./certs/
 