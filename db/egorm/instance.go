package egorm

import (
	"sync"

	"miopkg/log"
)

var instances = sync.Map{}

// iterate 遍历所有实例
func iterate(fn func(name string, db *Component) bool) {
	instances.Range(func(key, val interface{}) bool {
		return fn(key.(string), val.(*Component))
	})
}

// configs
func configs() map[string]interface{} {
	var rets = make(map[string]interface{})
	instances.Range(func(key, val interface{}) bool {
		return true
	})

	return rets
}

// stats
func stats() (stats map[string]interface{}) {
	stats = make(map[string]interface{})
	instances.Range(func(key, val interface{}) bool {
		name := key.(string)
		db := val.(*Component)

		sqlDB, err := db.DB()
		if err != nil {
			log.MioLogger.With(log.FieldMod(PackageName)).Panic("stats db error", log.FieldErr(err))
			return false
		}
		stats[name] = sqlDB.Stats()
		return true
	})

	return
}
