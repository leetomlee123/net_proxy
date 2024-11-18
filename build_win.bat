del app.exe
SET GOOS=windows
SET GOARCH=amd64
go mod tidy
go build -o app.exe