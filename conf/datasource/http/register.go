package http

import (
	"miopkg/conf"
	"miopkg/flag"
	"miopkg/log"
)

// Defines http/https scheme
const (
	DataSourceHttp  = "http"
	DataSourceHttps = "https"
)

func init() {
	dataSourceCreator := func() conf.DataSource {
		var (
			watchConfig = flag.Bool("watch")
			configAddr  = flag.String("config")
		)
		if configAddr == "" {
			log.Panic("new http dataSource, configAddr is empty")
			return nil
		}
		return NewDataSource(configAddr, watchConfig)
	}
	conf.Register(DataSourceHttp, dataSourceCreator)
	conf.Register(DataSourceHttps, dataSourceCreator)
}
