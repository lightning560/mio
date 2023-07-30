package flag

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/// FlagSet用于的注册解析调用。
/// Bool之类的方法单个flag的查询

var (
	flagset      *FlagSet
	defaultFlags = []Flag{
		// HelpFlag prints usage of application.
		/// 1 Flag接口的[]，创建struct已经实现Flag接口
		&BoolFlag{
			Name:  "help",
			Usage: "--help, show help information",
			Action: func(name string, fs *FlagSet) {
				/// 用于加载到自定义的flagset
				fs.PrintDefaults()
				/// PrintDefault会向标准错误输出写入所有注册好的flag的默认值。
				os.Exit(0)
				/// 0表示正常退出，其他数字都表示异常退出
			},
		},
	}
)

func init() {
	// procName := filepath.Base(os.Args[0])
	// nfs := flag.NewFlagSet(procName, flag.ExitOnError)
	///flag也有一个FlagSet，初始化一个FlagSet
	flagset = &FlagSet{
		FlagSet:  flag.CommandLine,
		flags:    defaultFlags,
		actions:  make(map[string]func(string, *FlagSet)),
		environs: make(map[string]string),
	}
}

// Flag ...
type (
	// Flag defines application flag.
	Flag interface {
		Apply(*FlagSet)
	}

	// FlagSet wraps a set of Flags.
	///主要是*flag.FlagSet在做事，flags的数据会被parse
	FlagSet struct {
		*flag.FlagSet                                   /// 用于调用实现
		flags         []Flag                            ///用于注册
		actions       map[string]func(string, *FlagSet) ///用于将flags中的flag解析并放入map
		environs      map[string]string                 /// 读取os.GetEnv,然后存入
	}
)

// Register ...
/// 1 和With一样。添加一个实现了apply接口的Flag
func Register(fs ...Flag) {
	flagset.Register(fs...)
}

// Register ...
func (fs *FlagSet) Register(flags ...Flag) {
	fs.flags = append(fs.flags, flags...)
}

// With adds flags to the flagset.
func With(fs ...Flag) { flagset.With(fs...) }

// With adds flags to provided flagset.
func (fs *FlagSet) With(flags ...Flag) {
	fs.flags = append(fs.flags, flags...)
}

// Parse parses the flagset.
/// 1 注册完成后，第一步，在初始化app时sync.once调用一次Parse方法
func Parse() error {
	return flagset.Parse()
}

// Lookup lookup flag value by name
// priority: flag > default > env
/// lookup的优先级flag > default > environs
func (fs *FlagSet) Lookup(name string) *flag.Flag {
	flag := fs.FlagSet.Lookup(name)
	if flag != nil {
		if flag.Value.String() == "" {
			if env, ok := fs.environs[name]; ok {
				_ = flag.Value.Set(env)
			}
		}
		if flag.Value.String() == "" {
			_ = flag.Value.Set(flag.DefValue)
		}
	}
	return flag
}

// Parse parses provided flagset.
/// 1
func (fs *FlagSet) Parse() error {
	/// 判断是否可解析
	if fs.Parsed() {
		return nil
	}
	/// 2 遍历出Flag.flags，然后执行f的apply方法
	for _, f := range fs.flags {
		f.Apply(fs) ///3
	}
	/// 4 仅仅解析，解析出启动时的flag从第二个开始
	if err := fs.FlagSet.Parse(os.Args[1:]); err != nil {
		return err
	}
	///5 遍历4中解析处理每个flag指令，然后取出对应的名字
	fs.FlagSet.Visit(func(f *flag.Flag) {
		// do action hook after parse flagset
		if action, ok := fs.actions[f.Name]; ok && action != nil {
			action(f.Name, fs)
		}
		if env, ok := fs.environs[f.Name]; ok {
			fs.environs[f.Name] = env
		}
	})

	return nil
}

// BoolFlag is a bool flag implements of Flag interface.
type BoolFlag struct {
	Name     string ///对应FlagSet的field
	Usage    string
	EnvVar   string
	Default  bool
	Variable *bool
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
/// 3 实现Apply的struct，传入struct，传出FlagSet。就是将flagset中的[]flag，解析到flagset的action和environs中
func (f *BoolFlag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			///flag的现成方法处理
			set.FlagSet.BoolVar(f.Variable, field, f.Default, f.Usage)
		}
		///flag的现成方法处理
		set.FlagSet.Bool(field, f.Default, f.Usage)
		set.actions[field] = f.Action
		set.environs[field] = os.Getenv(f.EnvVar)
	}
}

// StringFlag is a string flag implements of Flag interface.
type StringFlag struct {
	Name     string
	Usage    string
	EnvVar   string
	Default  string
	Variable *string
	// Action hooked after call fs.Parse()
	Action func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *StringFlag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.StringVar(f.Variable, field, f.Default, f.Usage)
		}
		set.FlagSet.String(field, f.Default, f.Usage)
		set.actions[field] = f.Action
		set.environs[field] = os.Getenv(f.EnvVar)
	}
}

// IntFlag is an int flag implements of Flag interface.
type IntFlag struct {
	Name     string
	Usage    string
	Default  int
	Variable *int
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *IntFlag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.IntVar(f.Variable, field, f.Default, f.Usage)
		}
		set.FlagSet.Int(field, f.Default, f.Usage)
		set.actions[field] = f.Action
	}
}

// UintFlag is an uint flag implements of Flag interface.
type UintFlag struct {
	Name     string
	Usage    string
	Default  uint
	Variable *uint
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *UintFlag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.UintVar(f.Variable, field, f.Default, f.Usage)
		}
		set.FlagSet.Uint(field, f.Default, f.Usage)
		set.actions[field] = f.Action
	}
}

// Float64Flag is a float flag implements of Flag interface.
type Float64Flag struct {
	Name     string
	Usage    string
	Default  float64
	Variable *float64
	Action   func(string, *FlagSet)
}

// Apply implements of Flag Apply function.
func (f *Float64Flag) Apply(set *FlagSet) {
	for _, field := range strings.Split(f.Name, ",") {
		field = strings.TrimSpace(field)
		if f.Variable != nil {
			set.FlagSet.Float64Var(f.Variable, field, f.Default, f.Usage)
		}
		set.FlagSet.Float64(field, f.Default, f.Usage)
		set.actions[field] = f.Action
	}
}

/// query func
/// 这些方法用来查询具体的name关联的flag输入，都是调用lookup
// BoolE parses bool flag of the flagset with error returned.
func BoolE(name string) (bool, error) { return flagset.BoolE(name) }

// BoolE parses bool flag of provided flagset with error returned.
/// 4个方法最后的执行就是BoolE，调用lookup找到具体的flag，最终执行的是flag.Value
func (fs *FlagSet) BoolE(name string) (bool, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return strconv.ParseBool(flag.Value.String())
	}

	return false, fmt.Errorf("undefined flag name: %s", name)
}

// Bool parses bool flag of the flagset.
func Bool(name string) bool { return flagset.Bool(name) }

// Bool parses bool flag of provided flagset.
func (fs *FlagSet) Bool(name string) bool {
	ret, _ := fs.BoolE(name)
	return ret
}

// StringE parses string flag of the flagset with error returned.
func StringE(name string) (string, error) { return flagset.StringE(name) }

// StringE parses string flag of provided flagset with error returned.
func (fs *FlagSet) StringE(name string) (string, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return flag.Value.String(), nil
	}

	return "", fmt.Errorf("undefined flag name: %s", name)
}

// String parses string flag of the flagset.
func String(name string) string { return flagset.String(name) }

// String parses string flag of provided flagset.
func (fs *FlagSet) String(name string) string {
	ret, _ := fs.StringE(name)
	return ret
}

// IntE parses int flag of the flagset with error returned.
func IntE(name string) (int64, error) { return flagset.IntE(name) }

// IntE parses int flag of provided flagset with error returned.
func (fs *FlagSet) IntE(name string) (int64, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return strconv.ParseInt(flag.Value.String(), 10, 64)
	}

	return 0, fmt.Errorf("undefined flag name: %s", name)
}

// Int parses int flag of the flagset.
func Int(name string) int64 { return flagset.Int(name) }

// Int parses int flag of provided flagset.
func (fs *FlagSet) Int(name string) int64 {
	ret, _ := fs.IntE(name)
	return ret
}

// UintE parses int flag of the flagset with error returned.
func UintE(name string) (uint64, error) { return flagset.UintE(name) }

// UintE parses int flag of provided flagset with error returned.
func (fs *FlagSet) UintE(name string) (uint64, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return strconv.ParseUint(flag.Value.String(), 10, 64)
	}

	return 0, fmt.Errorf("undefined flag name: %s", name)
}

// Uint parses int flag of the flagset.
func Uint(name string) uint64 { return flagset.Uint(name) }

// Uint parses int flag of provided flagset.
func (fs *FlagSet) Uint(name string) uint64 {
	ret, _ := fs.UintE(name)
	return ret
}

// Float64E parses int flag of the flagset with error returned.
func Float64E(name string) (float64, error) { return flagset.Float64E(name) }

// Float64E parses int flag of provided flagset with error returned.
func (fs *FlagSet) Float64E(name string) (float64, error) {
	flag := fs.Lookup(name)
	if flag != nil {
		return strconv.ParseFloat(flag.Value.String(), 64)
	}

	return 0.0, fmt.Errorf("undefined flag name: %s", name)
}

// Float64 parses int flag of the flagset.
func Float64(name string) float64 { return flagset.Float64(name) }

// Float64 parses int flag of provided flagset.
func (fs *FlagSet) Float64(name string) float64 {
	ret, _ := fs.Float64E(name)
	return ret
}
