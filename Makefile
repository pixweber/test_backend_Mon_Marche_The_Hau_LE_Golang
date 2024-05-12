test_load:
	siege -c 100 -t 1 -H 'Content-Type: text/plain' 'http://localhost:8080/ticket POST < payload.txt'