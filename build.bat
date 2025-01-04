@echo off

SET GOOS=windows

go build -ldflags="-s -w" -o dist/win/CEMCAU.exe src/main.go

SET GOOS=linux

go build -ldflags="-s -w" -o dist/linux/CEMCAU src/main.go