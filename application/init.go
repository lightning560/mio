package application

import (
	"miopkg/env"
	"miopkg/flag"
	"os"
)

func init() {
	flag.Register(&flag.BoolFlag{
		Name:    "version",
		Usage:   "--version, print version",
		Default: false,
		Action: func(string, *flag.FlagSet) {
			env.PrintVersion()
			os.Exit(0)
		},
	})
}
