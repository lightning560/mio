package egorm

import (
	"net/http"
	"time"

	"miopkg/governor"
	"miopkg/log"
	"miopkg/metric"

	jsoniter "github.com/json-iterator/go"
)

func init() {
	type gormStatus struct {
		Gorms map[string]interface{} `json:"gorms"`
	}
	var rets = gormStatus{
		Gorms: make(map[string]interface{}, 0),
	}
	governor.HandleFunc("/debug/gorm/stats", func(w http.ResponseWriter, r *http.Request) {
		rets.Gorms = stats()
		_ = jsoniter.NewEncoder(w).Encode(rets)
	})
	go monitor()
}

func monitor() {
	for {
		time.Sleep(time.Second * 10)
		iterate(func(name string, db *Component) bool {
			sqlDB, err := db.DB()
			if err != nil {
				log.MioLogger.With(log.FieldMod(PackageName)).Panic("monitor db error", log.FieldErr(err))
				return false
			}

			stats := sqlDB.Stats()
			metric.LibHandleSummary.Observe(float64(stats.Idle), name, "idle")
			metric.LibHandleSummary.Observe(float64(stats.InUse), name, "inuse")
			metric.LibHandleSummary.Observe(float64(stats.WaitCount), name, "wait")
			metric.LibHandleSummary.Observe(float64(stats.OpenConnections), name, "conns")
			metric.LibHandleSummary.Observe(float64(stats.MaxOpenConnections), name, "max_open_conns")
			metric.LibHandleSummary.Observe(float64(stats.MaxIdleClosed), name, "max_idle_closed")
			metric.LibHandleSummary.Observe(float64(stats.MaxLifetimeClosed), name, "max_lifetime_closed")
			return true
		})
	}
}
