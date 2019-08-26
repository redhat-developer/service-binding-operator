package collectors

type Collector interface {
	Collect() (BindableMetadata, error)
}

type BindableMetadata struct {
	BindableSecrets
	BindableVolumePaths
}

type BindableSecrets map[string][]byte

type BindableVolumePaths []string
