package expressions

import "fmt"

// UndefinedVariable is an error that the named variable is not defined.
type UndefinedVariable string

func (e UndefinedVariable) Error() string {
	return fmt.Sprintf("undefined variable %q", string(e))
}
