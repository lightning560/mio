package xmongo

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestObjectIdFromIntId(t *testing.T) {
	oid := primitive.NewObjectID()
	fmt.Println(oid)
}
func TestIntIdFromObjectId(t *testing.T) {
	oid := primitive.NewObjectID()
	int64Id, err := IntIdFromObjectId(oid)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(int64Id)
}
