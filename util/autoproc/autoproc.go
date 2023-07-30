package autoproc

import (
	"runtime"

	"miopkg/conf"
	"miopkg/errors"
	"miopkg/log"

	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// 初始化注册中心
	if _, err := maxprocs.Set(); err != nil {
		log.Panic("auto max procs", log.FieldMod(errors.ModProc), log.FieldErrKind(errors.ErrKindAny), log.FieldErr(err))
	}
	conf.OnLoaded(func(c *conf.Configuration) {
		if maxProcs := conf.GetInt("maxProc"); maxProcs != 0 {
			runtime.GOMAXPROCS(maxProcs)
		}
		log.Info("auto max procs", log.FieldMod(errors.ModProc), log.Int64("procs", int64(runtime.GOMAXPROCS(-1))))
	})
}
