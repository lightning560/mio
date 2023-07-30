package emongo

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type Database struct {
	mu        sync.Mutex
	db        *mongo.Database
	processor processor
	logMode   bool
}

func (wd *Database) Client() *Client {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	cc := wd.db.Client()
	if cc == nil {
		return nil
	}
	return &Client{cc: cc}
}

func (wd *Database) Collection(name string, opts ...*options.CollectionOptions) *Collection {
	if wd.db == nil {
		return nil
	}
	coll := wd.db.Collection(name, opts...)
	if coll == nil {
		return nil
	}
	return &Collection{coll: coll, processor: wd.processor, logMode: wd.logMode}
}

// Drop drops the database on the server.
// This method ignores "namespace not found" errors so it is safe to drop a database that does not exist on the server.
func (wd *Database) Drop(ctx context.Context) error {
	return wd.processor(func(c *cmd) error {
		logCmd(wd.logMode, c, "Drop", nil)
		return wd.db.Drop(ctx)
	})
}

func (wd *Database) ListCollections(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) (
	cur *mongo.Cursor, err error) {
	err = wd.processor(func(c *cmd) error {
		cur, err = wd.db.ListCollections(ctx, filter, opts...)
		logCmd(wd.logMode, c, "ListCollections", cur, filter)
		return err
	})
	return
}

func (wd *Database) Name() string                          { return wd.db.Name() }
func (wd *Database) ReadConcern() *readconcern.ReadConcern { return wd.db.ReadConcern() }
func (wd *Database) ReadPreference() *readpref.ReadPref    { return wd.db.ReadPreference() }

func (wd *Database) RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) (res *mongo.SingleResult) {
	_ = wd.processor(func(c *cmd) error {
		res = wd.db.RunCommand(ctx, runCommand, opts...)
		logCmd(wd.logMode, c, "RunCommand", res, runCommand)
		return res.Err()
	})
	return
}

func (wd *Database) WriteConcern() (res *writeconcern.WriteConcern) {
	_ = wd.processor(func(c *cmd) error {
		res = wd.db.WriteConcern()
		logCmd(wd.logMode, c, "WriteConcern", res)
		return nil
	})
	return
}

func (wd *Database) Database() *mongo.Database {
	return wd.db
}
