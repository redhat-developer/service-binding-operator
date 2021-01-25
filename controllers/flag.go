package controllers

import (
	"flag"
)

var (
	maxConcurrentReconciles int
)

func RegisterFlags(flags *flag.FlagSet) {
	flags.IntVar(&maxConcurrentReconciles, "max-concurrent-reconciles", 1, "max-concurrent-reconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 1.")
}
