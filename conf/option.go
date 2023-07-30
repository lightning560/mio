

package conf

// GetOption ...
type (
	GetOption  func(o *GetOptions)
	GetOptions struct {
		TagName   string
		Namespace string
		Module    string
	}
)

var defaultGetOptions = GetOptions{
	TagName: "mapstructure",
}

// 设置Tag
func TagName(tag string) GetOption {
	return func(o *GetOptions) {
		o.TagName = tag
	}
}

func TagNameJSON() GetOption {
	return TagName("json")
}

func TagNameTOML() GetOption {
	return TagName("toml")
}

func TagNameYAML() GetOption {
	return TagName("yaml")
}

func BuildinModule(module string) GetOption {
	return func(o *GetOptions) {
		o.Namespace = "mio"
		o.Module = module
	}
}

func Namespace(namespace string) GetOption {
	return func(o *GetOptions) {
		o.Namespace = namespace
	}
}

func Module(module string) GetOption {
	return func(o *GetOptions) {
		o.Module = module
	}
}
