package expressions

type UndefinedVariableHandler interface {
	OnUndefinedVariable(name string) interface{}
}

type UndefinedFilterHandler interface {
	OnUndefinedFilter(name string) interface{}
}

type StrictMode struct{}
type LaxMode struct{}

// Panics when a filter is missing.
func (mode StrictMode) OnUndefinedFilter(name string) interface{} {
	panic(UndefinedFilter(name))
}

// Panics when a variable is missing.
func (mode StrictMode) OnUndefinedVariable(name string) interface{} {
	panic(UndefinedVariable(name))
}

func identityFilter(i interface{}) interface{} {
	return i
}

// Uses an identity function as a default to ignore missing filters.
func (mode LaxMode) OnUndefinedFilter(name string) interface{} {
	return identityFilter
}

// Uses nil as a default value to ignore missing variables.
func (mode LaxMode) OnUndefinedVariable(name string) interface{} {
	return nil
}
