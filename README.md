# API for DynamoDB CRUD operations


 Project Structure
  ├── main.go # All implementation
  ├── go.mod
  └── go,sum


## Get all dependencies
```
go get github.com/aws/aws-lambda-go
```

## To build for Linux for windows
### cmd
```cmd
set GOOS=linux
go build -o main main.go
```

```powershell
$env:GOOS = "linux"
$env:CGO_ENABLED = "0"
$env:GOARCH = "amd64"
go build -o main main.go
```

## To create ZIP
### Get lambda zip tool
```
go.exe get -u github.com/aws/aws-lambda-go/cmd/build-lambda-zip
```
### To ZIP
#### cmd
``` cmd
%USERPROFILE%\Go\bin\build-lambda-zip.exe -output main.zip main
```

``` powershell
~\Go\Bin\build-lambda-zip.exe -output main.zip main
```
