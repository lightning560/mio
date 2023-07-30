package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"miopkg/conf"
	"miopkg/examples/db/emongo/model"

	"github.com/BurntSushi/toml"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"miopkg/db/emongo"
)

// go run main.go
//
//export MIO_MODE=dev && go run main.go
func main() {
	var stopCh = make(chan bool)
	// 假设你配置的toml如下所示
	config := `
[mongo]
	debug=true
	dsn="mongodb://root:password@127.0.0.1:27017/?authSource=admin&readPreference=primary&appname=MongoDB%20Compass&ssl=false"
`
	// 加载配置文件
	err := conf.LoadFromReader(strings.NewReader(config), toml.Unmarshal)
	if err != nil {
		panic("LoadFromReader fail," + err.Error())
	}

	// 初始化emongo组件
	cmp := emongo.Load("mongo").Build()
	coll := cmp.Client().Database("light").Collection("test")
	findOne(coll)
	// insertOne(coll)
	// insertOneModel(coll)
	// findOneModel(coll)
	// findOneAndUpdate(coll)
	stopCh <- true
}

func findOne(coll *emongo.Collection) {
	res := coll.FindOne(context.TODO(), bson.M{"rid": 888})
	var result bson.M
	err := res.Decode(&result)
	if err != nil {
		fmt.Println("findone err occurs", err)
	}
	fmt.Println("findone result is", result)
}
func insertOneModel(coll *emongo.Collection) {
	m := model.Like{
		Mid:  123,
		Oid:  234,
		Bid:  345,
		Sid:  45678,
		Like: 1,
	}
	rv, err := coll.InsertOne(context.TODO(), bson.M{"mid": m.Mid, "oid": m.Oid, "bid": m.Bid, "like": 1, "create_at": time.Now().Unix(), "updated_at": time.Now().Unix()})
	if err != nil {
		return
	}
	fmt.Println("AddLike,InsertOne succ:", rv)
}
func findOneModel(coll *emongo.Collection) {
	fmt.Println("findOneModel start")
	var res model.Like
	m := model.Like{
		Mid:  123,
		Oid:  234,
		Bid:  345,
		Sid:  45678,
		Like: 1,
	}
	fmt.Println("m:", m)
	err := coll.FindOne(context.TODO(), bson.M{"mid": m.Mid, "oid": m.Oid}).Decode(&res)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if err == mongo.ErrNoDocuments {
			fmt.Println("ErrNoDocuments:", err)
		}
		fmt.Println(err)
	}
	fmt.Println("findOneModel res:", res)
}
func findOneAndUpdate(coll *emongo.Collection) {
	fmt.Println("findOneAndUpdate start")
	var res model.Like
	m := model.Like{
		Mid:  123,
		Oid:  234,
		Bid:  345,
		Like: 1,
	}
	rv := coll.FindOneAndUpdate(context.Background(), bson.M{"mid": m.Mid, "oid": m.Oid, "bid": m.Bid}, bson.M{"$set": bson.M{"like": 1, "updated_at": time.Now().Unix() + 1}})
	if rv.Err() != nil {
		fmt.Println("FindOneAndUpdate err:", rv.Err())
		return
	}
	err := rv.Decode(&res)
	if err != nil {
		fmt.Println("FindOneAndUpdate Decode err")
		return
	}
	fmt.Println("FindOneAndUpdate res:", res)
}

func insertOne(coll *emongo.Collection) {

}
