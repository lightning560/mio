package conf

import (
	"log"

	"miopkg/flag"

	"github.com/BurntSushi/toml"
)

const DefaultEnvPrefix = "APP_"

func init() {
	flag.Register(&flag.StringFlag{Name: "envPrefix", Usage: "--envPrefix=APP_", Default: DefaultEnvPrefix, Action: func(key string, fs *flag.FlagSet) {
		var envPrefix = fs.String(key)
		defaultConfiguration.LoadEnvironments(envPrefix)
	}})

	flag.Register(&flag.StringFlag{Name: "config", Usage: "--config=config.toml", Action: func(key string, fs *flag.FlagSet) {
		var configAddr = fs.String(key)
		log.Printf("read config: %s", configAddr)
		datasource, err := NewDataSource(configAddr)
		if err != nil {
			log.Fatalf("build datasource[%s] failed: %v", configAddr, err)
		}
		if err := LoadFromDataSource(datasource, toml.Unmarshal); err != nil {
			log.Fatalf("load config from datasource[%s] failed: %v", configAddr, err)
		}
		log.Printf("load config from datasource[%s] completely!", configAddr)
	}})

	flag.Register(&flag.StringFlag{Name: "config-tag", Usage: "--config-tag=mapstructure", Default: "mapstructure", Action: func(key string, fs *flag.FlagSet) {
		defaultGetOptions.TagName = fs.String("config-tag")
	}})

	flag.Register(&flag.StringFlag{Name: "config-namespace", Usage: "--config-namespace=mio, 配置内建组件的默认命名空间, mio", Default: "mio", Action: func(key string, fs *flag.FlagSet) {
		defaultGetOptions.Namespace = fs.String("config-namespace")
	}})

	flag.Register(&flag.BoolFlag{Name: "watch", Usage: "--watch, watch config change event", Default: false, EnvVar: "MIO_CONFIG_WATCH", Action: func(key string, fs *flag.FlagSet) {
		log.Printf("load config watch: %v", fs.Bool(key))
	}})
}
