package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"miopkg/component"
	"miopkg/worker/job"

	"miopkg/conf"

	"github.com/BurntSushi/toml"

	//go-lint
	_ "miopkg/conf/datasource/file"
	_ "miopkg/conf/datasource/http"
	_ "miopkg/registry/etcdv3"

	"miopkg/cycle"
	"miopkg/errors"
	"miopkg/flag"
	"miopkg/log"
	"miopkg/registry"
	"miopkg/server"
	"miopkg/util/hooks"
	"miopkg/util/signals"
	"miopkg/util/xdebug"
	"miopkg/util/xgo"
	"miopkg/worker"

	"github.com/fatih/color"
	"golang.org/x/sync/errgroup"
)

// Application is the framework's instance, it contains the servers, workers, client and configuration settings.
// Create an instance of Application, by using &Application{}
type Application struct {
	cycle    *cycle.Cycle
	smu      *sync.RWMutex
	initOnce sync.Once
	// startupOnce  sync.Once
	stopOnce sync.Once
	servers  []server.Server
	workers  []worker.Worker
	jobs     map[string]job.Runner
	logger   *log.Logger
	// hooks        map[uint32]*xdefer.DeferStack
	configParser conf.Unmarshaller
	disableMap   map[Disable]bool
	HideBanner   bool
	stopped      chan struct{}
	components   []component.Component
}

// New create a new Application instance
func New(fns ...func() error) (*Application, error) {
	app := &Application{}
	if err := app.Startup(fns...); err != nil {
		return nil, err
	}
	return app, nil
}

func DefaultApp() *Application {
	app := &Application{}
	app.initialize()
	return app
}

// run hooks
func (app *Application) runHooks(stage hooks.Stage) {
	hooks.Do(stage)
}

// RegisterHooks register a stage Hook
func (app *Application) RegisterHooks(stage hooks.Stage, fns ...func()) {
	hooks.Register(stage, fns...)
}

// initialize application
func (app *Application) initialize() {
	app.initOnce.Do(func() {
		//assign
		app.cycle = cycle.NewCycle()
		app.smu = &sync.RWMutex{}
		app.servers = make([]server.Server, 0)
		app.workers = make([]worker.Worker, 0)
		app.jobs = make(map[string]job.Runner)
		app.logger = log.MioLogger
		app.configParser = toml.Unmarshal //和toml的接口一致
		app.disableMap = make(map[Disable]bool)
		app.stopped = make(chan struct{})
		app.components = make([]component.Component, 0)
		//private method

		_ = app.parseFlags()
		_ = app.printBanner()
	})
}

// // start up application
// // By default the startup composition is:
// // - parse config, watch, version flags
// // - load config
// // - init default biz logger, mio frame logger
// // - init procs
// func (app *Application) startup() (err error) {
// 	app.startupOnce.Do(func() {
// 		err = xgo.SerialUntilError(
// 			app.parseFlags,
// 			// app.printBanner,
// 			// app.loadConfig,
// 			// app.initLogger,
// 			// app.initMaxProcs,
// 			// app.initTracer,
// 			// app.initSentinel,
// 			// app.initGovernor,
// 		)()
// 	})
// 	return
// }

// Startup ..
func (app *Application) Startup(fns ...func() error) error {
	app.initialize()
	// if err := app.startup(); err != nil {
	// 	return err
	// }
	return xgo.SerialUntilError(fns...)()
}

// Defer ..
// Deprecated: use AfterStop instead
// func (app *Application) Defer(fns ...func() error) {
// 	app.AfterStop(fns...)
// }

// BeforeStop hook
// Deprecated: use RegisterHooks instead
// func (app *Application) BeforeStop(fns ...func() error) {
// 	app.RegisterHooks(StageBeforeStop, fns...)
// }

// AfterStop hook
// Deprecated: use RegisterHooks instead
// func (app *Application) AfterStop(fns ...func() error) {
// 	app.RegisterHooks(StageAfterStop, fns...)
// }

// Serve start server
func (app *Application) Serve(s ...server.Server) error {
	app.smu.Lock()
	defer app.smu.Unlock()
	app.servers = append(app.servers, s...)
	return nil
}

// Schedule ..
func (app *Application) Schedule(w worker.Worker) error {
	app.workers = append(app.workers, w)
	return nil
}

// Job ..
func (app *Application) Job(runner job.Runner) error {
	namedJob, ok := runner.(interface{ GetJobName() string })
	// job runner must implement GetJobName
	if !ok {
		return nil
	}
	jobName := namedJob.GetJobName()
	if flag.Bool("disable-job") {
		app.logger.Info("mio disable job", log.FieldName(jobName))
		return nil
	}

	// start job by name
	jobFlag := flag.String("job")
	if jobFlag == "" {
		app.logger.Error("mio jobs flag name empty", log.FieldName(jobName))
		return nil
	}

	if jobName != jobFlag {
		app.logger.Info("mio disable jobs", log.FieldName(jobName))
		return nil
	}
	app.logger.Info("mio register job", log.FieldName(jobName))
	app.jobs[jobName] = runner
	return nil
}

// SetRegistry set customize registry
// Deprecated, please use registry.DefaultRegisterer instead.
func (app *Application) SetRegistry(reg registry.Registry) {
	registry.DefaultRegisterer = reg
}

// SetGovernor set governor addr (default 127.0.0.1:0)
// Deprecated
//func (app *Application) SetGovernor(addr string) {
//	app.governorAddr = addr
//}

// Run run application
func (app *Application) Run(servers ...server.Server) error {
	app.smu.Lock()
	app.servers = append(app.servers, servers...)
	app.smu.Unlock()

	app.waitSignals() //start signal listen task in goroutine
	defer app.clean()

	// todo jobs not graceful
	_ = app.startJobs()

	// start servers and govern server
	app.cycle.Run(app.startServers)
	// start workers
	app.cycle.Run(app.startWorkers)

	//blocking and wait quit
	if err := <-app.cycle.Wait(); err != nil {
		app.logger.Error("mio shutdown with error", log.FieldMod(errors.ModApp), log.FieldErr(err))
		return err
	}
	app.logger.Info("shutdown mio, bye!", log.FieldMod(errors.ModApp))
	return nil
}

// clean after app quit
func (app *Application) clean() {
	_ = log.DefaultLogger.Flush()
	_ = log.MioLogger.Flush()
}

// Stop application immediately after necessary cleanup
func (app *Application) Stop() (err error) {
	app.stopOnce.Do(func() {
		app.stopped <- struct{}{}
		app.runHooks(hooks.Stage_BeforeStop)

		//stop servers
		app.smu.RLock()
		for _, s := range app.servers {
			func(s server.Server) {
				app.cycle.Run(s.Stop)
			}(s)
		}
		app.smu.RUnlock()

		//stop workers
		for _, w := range app.workers {
			func(w worker.Worker) {
				app.cycle.Run(w.Stop)
			}(w)
		}
		<-app.cycle.Done()
		app.runHooks(hooks.Stage_AfterStop)
		app.cycle.Close()
	})
	return
}

// GracefulStop application after necessary cleanup
func (app *Application) GracefulStop(ctx context.Context) (err error) {
	app.stopOnce.Do(func() {
		app.stopped <- struct{}{}
		app.runHooks(hooks.Stage_BeforeStop)

		//stop servers
		app.smu.RLock()
		for _, s := range app.servers {
			func(s server.Server) {
				app.cycle.Run(func() error {
					return s.GracefulStop(ctx)
				})
			}(s)
		}
		app.smu.RUnlock()

		//stop workers
		for _, w := range app.workers {
			func(w worker.Worker) {
				app.cycle.Run(w.Stop)
			}(w)
		}
		<-app.cycle.Done()
		app.runHooks(hooks.Stage_AfterStop)
		app.cycle.Close()
	})
	return err
}

// waitSignals wait signal
func (app *Application) waitSignals() {
	app.logger.Info("init listen signal", log.FieldMod(errors.ModApp), log.FieldEvent("init"))
	signals.Shutdown(func(grace bool) { //when get shutdown signal
		//todo: support timeout
		if grace {
			_ = app.GracefulStop(context.TODO())
		} else {
			_ = app.Stop()
		}
	})
}

// func (app *Application) initGovernor() error {
// 	if app.isDisable(DisableDefaultGovernor) {
// 		app.logger.Info("defualt governor disable", log.FieldMod(errors.ModApp))
// 		return nil
// 	}

// 	config := governor.StdConfig("governor")
// 	if !config.Enable {
// 		return nil
// 	}
// 	return app.Serve(config.Build())
// }

func (app *Application) startServers() error {
	var eg errgroup.Group
	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	go func() {
		<-app.stopped
		cancel()
	}()
	// start multi servers
	for _, s := range app.servers {
		s := s
		eg.Go(func() (err error) {
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_ = registry.DefaultRegisterer.UnregisterService(ctx, s.Info())
				app.logger.Info("exit server", log.FieldMod(errors.ModApp), log.FieldEvent("exit"), log.FieldName(s.Info().Name), log.FieldErr(err), log.FieldAddr(s.Info().Label()))
			}()

			time.AfterFunc(time.Second, func() {
				_ = registry.DefaultRegisterer.RegisterService(ctx, s.Info())
				app.logger.Info("start server", log.FieldMod(errors.ModApp), log.FieldEvent("init"), log.FieldName(s.Info().Name), log.FieldAddr(s.Info().Label()), log.Any("scheme", s.Info().Scheme))
			})
			err = s.Serve()
			return
		})
	}
	return eg.Wait()
}

func (app *Application) startWorkers() error {
	var eg errgroup.Group
	// start multi workers
	for _, w := range app.workers {
		w := w
		eg.Go(func() error {
			return w.Run()
		})
	}
	return eg.Wait()
}

// todo handle error
func (app *Application) startJobs() error {
	if len(app.jobs) == 0 {
		return nil
	}
	var jobs = make([]func(), 0)
	//warp jobs
	for name, runner := range app.jobs {
		jobs = append(jobs, func() {
			app.logger.Info("job run begin", log.FieldName(name))
			defer app.logger.Info("job run end", log.FieldName(name))
			// runner.Run panic 错误在更上层抛出
			runner.Run()
		})
	}
	xgo.Parallel(jobs...)()
	return nil
}

// parseFlags init
func (app *Application) parseFlags() error {
	if app.isDisable(DisableParserFlag) {
		app.logger.Info("parseFlags disable", log.FieldMod(errors.ModApp))
		return nil
	}
	// flag.Register(&flag.StringFlag{
	// 	Name:    "config",
	// 	Usage:   "--config",
	// 	EnvVar:  "Mio_CONFIG",
	// 	Default: "",
	// 	Action:  func(name string, fs *flag.FlagSet) {},
	// })

	// flag.Register(&flag.BoolFlag{
	// 	Name:    "version",
	// 	Usage:   "--version, print version",
	// 	Default: false,
	// 	Action: func(string, *flag.FlagSet) {
	// 		pkg.PrintVersion()
	// 		os.Exit(0)
	// 	},
	// })

	// flag.Register(&flag.StringFlag{
	// 	Name:    "host",
	// 	Usage:   "--host, print host",
	// 	Default: "127.0.0.1",
	// 	Action:  func(string, *flag.FlagSet) {},
	// })
	return flag.Parse()
}

//loadConfig init
// func (app *Application) loadConfig() error {
// 	if app.isDisable(DisableLoadConfig) {
// 		app.logger.Info("load config disable", log.FieldMod(errors.ModConfig))
// 		return nil
// 	}

// 	var configAddr = flag.String("config")
// 	provider, err := manager.NewDataSource(configAddr)
// 	if err != manager.ErrConfigAddr {
// 		if err != nil {
// 			app.logger.Panic("data source: provider error", log.FieldMod(errors.ModConfig), log.FieldErr(err))
// 		}

// 		if err := conf.LoadFromDataSource(provider, app.configParser); err != nil {
// 			app.logger.Panic("data source: load config", log.FieldMod(errors.ModConfig), log.FieldErrKind(errors.ErrKindUnmarshalConfigErr), log.FieldErr(err))
// 		}
// 	} else {
// 		app.logger.Info("no config... ", log.FieldMod(errors.ModConfig))
// 	}
// 	return nil
// }

//initLogger init
// func (app *Application) initLogger() error {
// 	if conf.Get(log.ConfigEntry("default")) != nil {
// 		log.DefaultLogger = log.RawConfig(constant.ConfigPrefix + ".logger.default").Build()
// 	}
// 	log.DefaultLogger.AutoLevel(constant.ConfigPrefix + ".logger.default")

// 	if conf.Get(constant.ConfigPrefix+".logger.mio") != nil {
// 		log.MioLogger = log.RawConfig(constant.ConfigPrefix + ".logger.mio").Build()
// 	}
// 	log.MioLogger.AutoLevel(constant.ConfigPrefix + ".logger.mio")

// 	return nil
// }

//initTracer init
// func (app *Application) initTracer() error {
// 	// init tracing component jaeger
// 	if conf.Get("mio.trace.jaeger") != nil {
// 		var config = jaeger.RawConfig("mio.trace.jaeger")
// 		trace.SetGlobalTracer(config.Build())
// 	}
// 	return nil
// }

//initSentinel init
// func (app *Application) initSentinel() error {
// 	// init reliability component sentinel
// 	if conf.Get("mio.reliability.sentinel") != nil {
// 		app.logger.Info("init sentinel")
// 		return sentinel.RawConfig("mio.reliability.sentinel").Build()
// 	}
// 	return nil
// }

//initMaxProcs init
// func (app *Application) initMaxProcs() error {
// 	if maxProcs := conf.GetInt("maxProc"); maxProcs != 0 {
// 		runtime.GOMAXPROCS(maxProcs)
// 	} else {
// 		if _, err := maxprocs.Set(); err != nil {
// 			app.logger.Panic("auto max procs", log.FieldMod(errors.ModProc), log.FieldErrKind(errors.ErrKindAny), log.FieldErr(err))
// 		}
// 	}
// 	app.logger.Info("auto max procs", log.FieldMod(errors.ModProc), log.Int64("procs", int64(runtime.GOMAXPROCS(-1))))
// 	return nil
// }

func (app *Application) isDisable(d Disable) bool {
	b, ok := app.disableMap[d]
	if !ok {
		return false
	}
	return b
}

// printBanner init
func (app *Application) printBanner() error {
	if app.HideBanner {
		return nil
	}

	if xdebug.IsTestingMode() {
		return nil
	}
	// http://www.figlet.org/examples.html
	const banner = `
    _/      _/    _/                 
   _/_/  _/_/    _/      _/_/    
  _/  _/  _/    _/    _/    _/   
 _/      _/    _/    _/    _/    
_/      _/    _/      _/_/       

Welcome to miopkg, starting application ...
`
	fmt.Println(color.GreenString(banner))
	return nil
}
