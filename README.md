# maiyajia-api-server

## 麦芽＋创新教育平台API服务，包括功能：

- 学生/教师注册登录
- 学生签到
- 学生/教师管理
- 班级管理
- 用户消息
- 软件工具
- 课程商场
- 作品管理
- 学习进度管理
- 社区平台


## 安装

- (1) git clone 工程到本机的golang src 目录下，更名为maiyajia.com

` git clone http://192.168.2.100:3000/maiyajia-next/maiyajia-api-server.git maiyajia.com `

- (2) cd到maiyajia.com目录，运行下面命令

` bee run `

## 文档

 [API接口文档](http://192.168.2.100:3000/maiyajia-next/maiyajia-api-server-doc)

 



## 内网部署说明

1、说明
- （1）内网IP 192.168.2.100

- （2）端口 8888

- （3）启动服务 sudo systemctl start maiyajia.com

- （4)停止服务 sudo systemctl stop maiyajia.com


2、编译

    cd ~/golang-workspace/scr/maiyajia.com

    sudo systemctl stop maiyajia.com

    bee run 

    Ctrl + c  // 停止运行

    sudo systemctl start maiyajia.com

## 服务的启动与停止
- 启动服务 startServer.bat（右键->以管理员身份运行）
- 停止服务 startServer.bat（右键->以管理员身份运行）
