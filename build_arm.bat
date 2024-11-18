del net.so
go mod tidy
set GOARCH=arm64
set GOOS=linux
go build -o net.so
