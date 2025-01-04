@echo off

go build -ldflags="-s -w" -o dist/win/CEMCAU.exe src/main.go