package models

import (
	"errors"
	"time"

	"github.com/astaxie/beego"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
)

type MessageModels struct {
	MgoSession *mongo.MgoClient
}

// MessageCode 消息类型代码
type MessageCode int

const (

	// SYSTEM 系统消息类型
	SYSTEM MessageCode = iota
	// CLASS 班级消息类型
	CLASS

	// 已读和未读消息类型
	READ   MessageCode = 0 // 已读
	UNREAD MessageCode = 1 // 未读
)

// Message 消息体
type Message struct {
	ID   bson.ObjectId `bson:"_id" json:"id"`
	Code MessageCode   `bson:"code" json:"code"`
	/**
	* 关于Publisher发布者做如下说明：
	* 系统消息的发布者，统一设置为 publisher := bson.ObjectIdHex("000000000000000000000000")
	* 班级消息的发布者，统一设置为班级ID
	* 以上约定，在客户端读取消息的时候，要注意进行区别。
	**/
	Publisher  bson.ObjectId `bson:"publisher" json:"publisher"`
	Title      string        `bson:"title" json:"title"`
	Content    string        `bson:"Content" json:"Content"`       // 消息内容
	CreateTime time.Time     `bson:"createTime" json:"createTime"` //创建时间
	ClassName  string        `bson:"className" json:"className"`
	Logo       string        `bson:"logo" json:"logo"`
}

// MessageQueue 消息队列
type MessageQueue struct {
	ID         bson.ObjectId `bson:"_id" json:"id"`
	Subscriber bson.ObjectId `bson:"subscriber" json:"subscriber"`
	MessageID  bson.ObjectId `bson:"messageID" json:"messageID"`
	Status     MessageCode   `bson:"status" json:"status"`
}

// NewMessage 创建一条新的消息
func NewMessage(code MessageCode, publisher bson.ObjectId, title, content, className, logo string) *Message {
	message := &Message{
		ID:         bson.NewObjectId(),
		Code:       code,
		Publisher:  publisher,
		Title:      title,
		Content:    content,
		CreateTime: time.Now(),
		ClassName:  className,
		Logo:       logo,
	}
	return message
}

// PublishSystemMessage 发布系统消息
func (messageMol *MessageModels) PublishSystemMessage(title, content, className, logo string, subscribers []bson.ObjectId) error {

	code := SYSTEM
	// 系统消息的发布者，统一设置为 publisher := bson.ObjectIdHex("000000000000000000000000")
	publisher := bson.ObjectIdHex("000000000000000000000000")
	message := NewMessage(code, publisher, title, content, className, logo)

	// 把消息体存库
	f := func(col *mgo.Collection) error {
		return col.Insert(message)
	}
	if err := messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message", f); err != nil {
		return err
	}

	// 创建消息队列
	var queues []interface{}
	for _, v := range subscribers {
		mq := &MessageQueue{
			ID:         bson.NewObjectId(),
			Subscriber: v,
			MessageID:  message.ID,
			Status:     UNREAD,
		}
		queues = append(queues, mq)
	}

	// 把消息队列存库
	ff := func(col *mgo.Collection) error {
		return col.Insert(queues...)
	}

	return messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message_queue", ff)
}

// PublishClassMessage 发布班级消息（用户教师或管理员向指定班级发布消息）
func (messageMol *MessageModels) PublishClassMessage(title, content, className, logo string, classID, userID bson.ObjectId) error {

	code := CLASS
	publisher := classID // 注意：如果是班级信息，则发布人设置为班级ID
	message := NewMessage(code, publisher, title, content, className, logo)
	// 把消息体存库
	f1 := func(col *mgo.Collection) error {
		return col.Insert(message)
	}
	if err := messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message", f1); err != nil {
		beego.Informational("insert message error")
		return err
	}

	// 获取班级的学生，并把学生ID设置为订阅者
	result := struct {
		Students []*Classmate `bson:"students"`
	}{}
	f2 := func(col *mgo.Collection) error {
		return col.FindId(classID).Select(bson.M{"students": 1}).One(&result)
	}
	if err := messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f2); err != nil {
		beego.Informational("find students error")
		return err
	}
	if len(result.Students) == 0 {
		return errors.New("not student")
	}

	// 把消息队列存库
	f3 := func(col *mgo.Collection) error {
		bulk := col.Bulk()
		bulk.Unordered()
		// 创建队列
		var queues []interface{}
		for _, v := range result.Students {
			mq := &MessageQueue{
				ID:         bson.NewObjectId(),
				Subscriber: v.UserID,
				MessageID:  message.ID,
				Status:     UNREAD,
			}
			queues = append(queues, mq)
		}
		//发送者ID也加入消息队列中
		mq := &MessageQueue{
			ID:         bson.NewObjectId(),
			Subscriber: userID,
			MessageID:  message.ID,
			Status:     READ,
		}
		queues = append(queues, mq)
		// 批量插入
		bulk.Insert(queues...)
		_, err := bulk.Run()
		return err
	}
	return messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message_queue", f3)

}

// ReadMessage 未读信息被读取，即从消息队列中把消息的IsRead字段设置为true
func (messageMol *MessageModels) ReadMessage(reader, messageID bson.ObjectId) error {

	query := bson.M{"subscriber": reader, "messageID": messageID}

	f := func(col *mgo.Collection) error {
		return col.Update(query, bson.M{"$set": bson.M{"status": READ}})
	}

	return messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message_queue", f)
}

// QueryUnreadMessageCount 查询订阅消息用户有多少条新的消息
func (messageMol *MessageModels) QueryUnreadMessageCount(subscriber bson.ObjectId) (int, error) {

	query := bson.M{"subscriber": subscriber, "status": UNREAD}
	var count int
	var err error
	f := func(col *mgo.Collection) error {
		count, err = col.Find(query).Count()
		return err
	}

	return count, messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message_queue", f)
}

// QueryMessageCount 查询用户消息数量（包括消息总数和未读消息总数）
func (messageMol *MessageModels) QueryMessageCount(subscriber bson.ObjectId) (interface{}, error) {
	beego.Informational(READ)
	beego.Informational(UNREAD)
	pipeline := []bson.M{
		{"$match": bson.M{"subscriber": subscriber}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}, "unread": bson.M{"$sum": "$status"}}},
		{"$project": bson.M{"_id": 0}},
	}

	result := bson.M{}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&result)
	}
	return &result, messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message_queue", f)
}

// FetchMessages 查询用户订阅的信息， 分页返回
func (messageMol *MessageModels) FetchMessages(subscriber bson.ObjectId, paging PagingInfo) (interface{}, error) {

	offset := paging.Offset()
	limit := paging.Limit()

	result := []bson.M{}

	pipeline := []bson.M{
		{"$match": bson.M{"subscriber": subscriber}},
		{"$lookup": bson.M{
			"from":         "message",
			"foreignField": "_id",
			"localField":   "messageID",
			"as":           "message",
		}},
		{"$sort": bson.M{"messages.createTime": -1}},
		{"$skip": offset},
		{"$limit": limit},
		{"$unwind": "$message"},
		{"$project": bson.M{"_id": 0, "message._id": 0}},
	}

	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).All(&result)
	}
	return &result, messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message_queue", f)

}

// DeleteUserMessage 删除用户的消息
func (messageMol *MessageModels) DeleteUserMessage(subscriber, messageID bson.ObjectId) error {
	f := func(col *mgo.Collection) error {
		return col.Remove(bson.M{"subscriber": subscriber, "messageID": messageID})
	}
	return messageMol.MgoSession.Do(beego.AppConfig.String("MongoDB"), "message_queue", f)
}
