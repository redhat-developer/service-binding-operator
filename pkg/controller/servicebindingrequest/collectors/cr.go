package collectors

// CrCollector reads all data referred in plan instance
type CrCollector struct {
	ctx           context.Context // request context
	client        client.Client   // Kubernetes API client
}

// NewCSVCollector creates a new CSVCollector
func NewCrCollector(ctx context.Context, client client.Client) CrCollector {
	return CrCollector{
		ctx:           ctx,
		client:        client,
}

// Collect returns all bindable metadata
func (c *CrCollector) Collect() (*BindableMetadata, error) {
	/*
		Read from the CR JSON path
	*/
	return nil, nil
}