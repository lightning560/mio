package emongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientEncryption struct {
	cc        *mongo.ClientEncryption
	processor processor
	logMode   bool
}

func (wc *Client) NewClientEncryption(opts ...*options.ClientEncryptionOptions) (*ClientEncryption, error) {
	client, err := mongo.NewClientEncryption(wc.Client(), opts...)
	if err != nil {
		return nil, err
	}
	return &ClientEncryption{cc: client, processor: defaultProcessor, logMode: wc.logMode}, nil
}

func (wce *ClientEncryption) CreateDataKey(ctx context.Context, kmsProvider string, opts ...*options.DataKeyOptions) (
	id primitive.Binary, err error) {

	err = wce.processor(func(c *cmd) error {
		id, err = wce.cc.CreateDataKey(ctx, kmsProvider, opts...)
		logCmd(wce.logMode, c, "CreateDataKey", id)
		return err
	})
	return
}

func (wce *ClientEncryption) Encrypt(ctx context.Context, val bson.RawValue, opts ...*options.EncryptOptions) (
	value primitive.Binary, err error) {

	err = wce.processor(func(c *cmd) error {
		value, err = wce.cc.Encrypt(ctx, val, opts...)
		logCmd(wce.logMode, c, "Encrypt", value, val)
		return err
	})
	return
}

func (wce *ClientEncryption) Decrypt(ctx context.Context, val primitive.Binary) (value bson.RawValue, err error) {
	err = wce.processor(func(c *cmd) error {
		value, err = wce.cc.Decrypt(ctx, val)
		logCmd(wce.logMode, c, "Decrypt", value, val)
		return err
	})
	return
}

func (wce *ClientEncryption) Close(ctx context.Context) error {
	return wce.processor(func(c *cmd) error {
		logCmd(wce.logMode, c, "Close", nil)
		return wce.cc.Close(ctx)
	})
}
