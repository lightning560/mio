package file

import (
	"miopkg/conf"
	"miopkg/flag"
	"miopkg/log"
)

// DataSourceFile defines file scheme
const DataSourceFile = "file"

func init() {
	// 根据flag的参数生成相应的DataSource并且加入到map
	conf.Register(DataSourceFile, func() conf.DataSource {
		var (
			watchConfig = flag.Bool("watch")
			configAddr  = flag.String("config")
		)
		if configAddr == "" {
			log.Panic("new file dataSource, configAddr is empty")
			return nil
		}
		return NewDataSource(configAddr, watchConfig)
	})
}
