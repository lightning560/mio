package governor

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"miopkg/conf"
	"miopkg/env"
	"miopkg/util/xstring"

	jsoniter "github.com/json-iterator/go"
)

func init() {
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Print("hook config, init runtime(governor)")
	})
	registerHandlers()
}

func registerHandlers() {
	HandleFunc("/configs", func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		if r.URL.Query().Get("pretty") == "true" {
			encoder.SetIndent("", "    ")
		}
		_ = encoder.Encode(conf.Traverse("."))
	})

	HandleFunc("/debug/config", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(xstring.PrettyJSONBytes(conf.Traverse(".")))
	})

	HandleFunc("/debug/env", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = jsoniter.NewEncoder(w).Encode(os.Environ())
	})

	HandleFunc("/build/info", func(w http.ResponseWriter, r *http.Request) {
		serverStats := map[string]string{
			"name":       env.Name(),
			"appID":      env.AppID(),
			"appMode":    env.AppMode(),
			"appVersion": env.AppVersion(),
			"mioVersion": env.MioVersion(),
			"buildUser":  env.BuildUser(),
			"buildHost":  env.BuildHost(),
			"buildTime":  env.BuildTime(),
			"startTime":  env.StartTime(),
			"hostName":   env.HostName(),
			"goVersion":  env.GoVersion(),
		}
		_ = jsoniter.NewEncoder(w).Encode(serverStats)
	})
}
