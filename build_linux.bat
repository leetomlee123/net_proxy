del app
go mod tidy
SET GOOS=linux
SET GOARCH=amd64
go build app.go