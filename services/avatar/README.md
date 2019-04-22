# 自动产生像素风格头像

## 使用

```go
    img, err := govatar.Buidler(govatar.MALE, "随机数")
````

## 生成新的资源绑定文件bindata.go

需要安装 

```
go get -u github.com/jteeuwen/go-bindata/...

```

在当前目录下运行

```
go-bindata -nomemcopy -pkg bindata -o ./bindata/bindata.go -ignore "(.+)\.go" data/...

```


#### 参考
https://github.com/jteeuwen/go-bindata

https://github.com/o1egl/govatar