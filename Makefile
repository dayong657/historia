GO=go

all: main

main: windows_dist darwin_dist linux_dist


windows_dist: deps
	mkdir -p build/windows
	GOOS=windows GOARCH=amd64 go build -o build/windows/server.exe examples/server2.go
	GOOS=windows GOARCH=amd64 go build -o build/windows/hammer.exe examples/hammer.go
	
	
darwin_dist: deps
	mkdir -p build/darwin
	GOOS=darwin GOARCH=amd64 go build -o build/darwin/server examples/server2.go
	GOOS=darwin GOARCH=amd64 go build -o build/darwin/hammer examples/hammer.go
	
linux_dist: deps
	mkdir -p build/linux
	GOOS=linux GOARCH=amd64 go build -o build/linux/server examples/server2.go
	GOOS=linux GOARCH=amd64 go build -o build/linux/hammer examples/hammer.go

clean:
	rm -rf build

deps:
	go get github.com/gorilla/mux
	go get github.com/gorilla/context
	go get github.com/josephlewis42/historia

