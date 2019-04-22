## 升级命令行工具

升级命令行工具，当系统下载完成升级文件后，这个工具能热启动服务器。


## 交叉编译升级命令行工具

### Mac 下编译 Linux 和 Windows 

- (1) 64位可执行程序

    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build upgrade.go
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build upgrade.go

- (2) 32位可执行程序

    CGO_ENABLED=0 GOOS=linux GOARCH=386 go build upgrade.go
    CGO_ENABLED=0 GOOS=windows GOARCH=386 go build upgrade.go

### Linux 下编译 Mac 和 Windows 

- (1) 64位可执行程序

    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build upgrade.go
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build upgrade.go

-(2) 32位可执行程序

### Windows 下编译 Mac 和 Linux 

- (1) 64位可执行程序

    SET CGO_ENABLED=0
    SET GOOS=darwin
    SET GOARCH=amd64
    go build upgrade.go

    SET CGO_ENABLED=0
    SET GOOS=linux
    SET GOARCH=amd64
    go build upgrade.go

- (2)32位可执行程序

    SET CGO_ENABLED=0
    SET GOOS=darwin
    SET GOARCH=386
    go build upgrade.go

    SET CGO_ENABLED=0
    SET GOOS=linux
    SET GOARCH=386
    go build upgrade.go

GOOS：目标平台的操作系统（darwin、freebsd、linux、windows） 
GOARCH：目标平台的体系架构（386、amd64、arm） 
