package xmongo

import (
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func IntIdFromObjectId(oid primitive.ObjectID) (int64, error) {
	hexId := oid.Hex()
	//FIXME: strconv.ParseInt: parsing "645111198d24dd1dd2f5361d": value out of range
	// 转为10进制后26位
	int64Id, err := strconv.ParseInt(hexId, 16, 64)
	if err != nil {
		return -1, err
	}
	return int64Id, nil
}
func ObjectIdFromIntId(id int64) (primitive.ObjectID, error) {
	hexId := strconv.FormatInt(id, 16)
	oid, err := primitive.ObjectIDFromHex(hexId)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return oid, nil
}
