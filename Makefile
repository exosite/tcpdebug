all: linux osx windows

linux: build/linux-amd64/tcpdebug

osx: build/osx-amd64/tcpdebug

windows: build/win-amd64/tcpdebug.exe

# Linux Build
build/linux-amd64/tcpdebug: main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ github.com/exosite/tcpdebug
# OS X Build
build/osx-amd64/tcpdebug: main.go
	GOOS=darwin GOARCH=amd64 go build -o $@ github.com/exosite/tcpdebug
# Windows Build
build/win-amd64/tcpdebug.exe: main.go
	GOOS=windows GOARCH=amd64 go build -o $@ github.com/exosite/tcpdebug

clean:
	rm -f build/linux-amd64/tcpdebug
	rm -f build/osx-amd64/tcpdebug
	rm -f build/win-amd64/tcpdebug.exe
	rm -f *~

.PHONY: all clean linux osx windows
