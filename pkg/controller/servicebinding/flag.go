package servicebinding

import "github.com/spf13/pflag"

var (
	sboFlagSet              *pflag.FlagSet
	maxConcurrentReconciles int
)

func init() {
	sboFlagSet = pflag.NewFlagSet("sbo", pflag.ExitOnError)
	sboFlagSet.IntVar(&maxConcurrentReconciles, "max-concurrent-reconciles", 1, "max-concurrent-reconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 1.")
}

func FlagSet() *pflag.FlagSet {
	return sboFlagSet
}
