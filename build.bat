@echo off

SET GOOS=windows

go build -ldflags="-s -w" -o dist/win/CEMCAU-win32-amd64.exe src/main.go

SET GOOS=linux

go build -ldflags="-s -w" -o dist/linux/CEMCAU-linux-amd64 src/main.go

SET GOARCH=arm64

go build -ldflags="-s -w" -o dist/linux/CEMCAU-linux-arm64 src/main.go

SET GOOS=darwin

go build -ldflags="-s -w" -o dist/darwin/CEMCAU-darwin-arm64 src/main.go

SET GOOS=windows

go build -ldflags="-s -w" -o dist/win/CEMCAU-win32-arm64.exe src/main.go