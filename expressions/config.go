package expressions

// Config holds configuration information for expression interpretation.
type Config struct {
	filters map[string]interface{}

	FilterErrorMode   UndefinedFilterHandler
	VariableErrorMode UndefinedVariableHandler
}

// NewConfig creates a new Config.
func NewConfig() Config {
	return Config{
		FilterErrorMode:   StrictMode{},
		VariableErrorMode: LaxMode{},
	}
}

func (c *Config) GetFilter(name string) interface{} {
	if val, ok := c.filters[name]; ok {
		return val
	}
	return c.FilterErrorMode.OnUndefinedFilter(name)
}

func (c *Config) GetVariable(bindings map[string]interface{}, name string) interface{} {
	if val, ok := bindings[name]; ok {
		return val
	}
	return c.VariableErrorMode.OnUndefinedVariable(name)
}
