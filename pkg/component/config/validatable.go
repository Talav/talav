package config

// Validatable is implemented by configuration structs that enforce invariants
// after load (e.g. ranges, required fields, or consistency checks unmarshaling cannot express).
//
// Use a pointer receiver:
//
//	func (c *MyConfig) Validate() error { ... }
//
// When using fxconfig.AsConfig, use struct type T (for example MyConfig{}), not *MyConfig.
type Validatable interface {
	Validate() error
}

// Validate runs v.Validate when ptr is non-nil and *T implements [Validatable]; otherwise it returns nil.
// If ptr is nil, Validate returns nil without panicking.
func Validate[T any](ptr *T) error {
	if ptr == nil {
		return nil
	}

	v, ok := any(ptr).(Validatable)
	if !ok {
		return nil
	}

	return v.Validate()
}
