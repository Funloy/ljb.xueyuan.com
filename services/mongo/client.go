// @APIVersion 1.0.0
// @Title 数据库操作客户端
// @Description mongodb数据库操作客户端
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package mongo

import (
	mgo "gopkg.in/mgo.v2"
)

// DBClient 数据库客户端实例
var Client *MgoClient

// MgoClient 数据库服务客户端
type MgoClient struct {
	//连接会话
	MongoSession *mgo.Session
}

// StartSession 获取数据库会话
func (mc *MgoClient) StartSession() (err error) {
	mc.MongoSession, err = CopyMonotonicSession()
	return err
}

// CloseSession 关闭会话
func (mc *MgoClient) CloseSession() (err error) {

	if mc.MongoSession != nil {
		CloseSession(mc.MongoSession)
		mc.MongoSession = nil
	}
	return err
}

/* Do 执行数据库指令
* 参数: 数据库名，文档名，执行命令回调函数
* 返回: 执行错误
 */
func (mc *MgoClient) Do(databaseName, collectionName string, dbCall DBCall) (err error) {
	return Execute(mc.MongoSession, databaseName, collectionName, dbCall)
}
