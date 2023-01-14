build:
	go build ./cmd/main.go
	mv ./main ./oko
	chmod +x ./oko
	./oko version