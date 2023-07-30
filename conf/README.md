# init

注册4个flag

- config
- envPrefix
- config-tag
- config-namespace

启动的方式conf是注册到flag中的config的action
读取的value是configAddr

## query

Get
最终调用的是
`func (c *Configuration) find(key string) interface{}`

# 加载配置格式

### 从字符串中加载配置

```golang
var content = `[app] mode="dev"`
if err := conf.Load(bytes.NewBufferString(content), toml.Unmarshal); err != nil {
    panic(err)
}
```

### 从配置文件中加载配置

```golang
import (
    file_datasource "miopkg/datasource/file"
)

provider := file_datasource.NewDataSource(path)
if err := conf.Load(provider, toml.Unmarshal); err != nil {
    panic(err)
}
```

### 从etcd中加载配置

```golang
import (
   etcdv3_datasource "miopkg/datasource/etcdv3"
   "miopkg/client/etcdv3"
)
provider := etcdv3_datasource.NewDataSource(
    etcdv3.StdConfig("config_datasource").Build(),
    "/config/my-application",
)
if err := conf.Load(provider, json.Unmarshal); err != nil {
    panic(err)
}
```
