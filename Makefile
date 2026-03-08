.PHONY: run-demo build-hotreload clean

run-demo:
	go run ./cmd/hotreload --root ./testserver --build "go build -o ./bin/server ./cmd/server" --exec "./bin/server"

build-hotreload:
	go build -o ./bin/hotreload ./cmd/hotreload

clean:
	-rm -rf ./bin ./testserver/bin
