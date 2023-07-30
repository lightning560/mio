package job

import (
	"miopkg/flag"
)

func init() {
	flag.Register(
		&flag.StringFlag{
			Name:    "job",
			Usage:   "--job",
			Default: "",
		},
	)
}

// Runner ...
type Runner interface {
	Run()
}
