package servicebindingrequest

type CustomEnvPath struct {
	Json string
	Variables []string
	Paths map[string]string
}

func (c *CustomEnvPath) Parse()  {
	values := make(map[string]string)
	for _, v := range c.Variables {
		if c.IsPresentInPaths(v) {

		}
	}
}

func (c CustomEnvPath) IsPresentInPaths(v string) bool {
	for customEnv := range c.Paths {
		if customEnv == v {
			return true
		}
		continue
	}
	return false
}
