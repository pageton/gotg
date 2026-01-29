package functions

import (
	"fmt"
	"reflect"
)

// GetOpt returns the first option from a variadic parameter list along with a boolean indicating validity.
//
// This function handles optional parameters in a variadic way, following Go idioms:
//   - Returns (value, true) if a valid option is provided
//   - Returns (zero, false) if no option or invalid option (zero value or nil pointer) is provided
//   - Panics if more than one option is provided (misuse prevention)
//
// Valid option means:
//   - Not the zero value for the type T
//   - Not a nil pointer (when T is a pointer type)
//
// Type parameter T must be comparable (supports == and != operators).
//
// Parameters:
//   - opts: Variadic options (0 or 1 expected, panics on >1)
//
// Returns:
//   - T: The option value if valid, otherwise zero value
//   - bool: True if option is valid, false otherwise
//
// Example:
//
//	timeout, ok := GetOpt[time.Duration](opts...)
//	if ok {
//	    // use timeout
//	} else {
//	    // use default timeout
//	}
//
// Example with pointer type:
//
//	cfg, ok := GetOpt[*Config](opts...)
//	if ok {
//	    // use cfg (non-nil)
//	} else {
//	    // use default config
//	}
func GetOpt[T comparable](opts ...T) (T, bool) {
	if len(opts) == 0 {
		var zero T
		return zero, false
	}

	if len(opts) > 1 {
		panic(fmt.Sprintf("too many options: expected 0 or 1, got %d", len(opts)))
	}

	first := opts[0]
	if !validOpt(first) {
		var zero T
		return zero, false
	}

	return first, true
}

// GetOptDef returns the first option from a variadic parameter list, or the default value if none provided.
//
// This function is a convenience wrapper around GetOpt for cases where you always want a value.
// It returns the default value when:
//   - No option is provided
//   - The provided option is invalid (zero value or nil pointer)
//
// Type parameter T must be comparable (supports == and != operators).
//
// Parameters:
//   - def: Default value to return if no valid option is provided
//   - opts: Variadic options (0 or 1 expected, panics on >1)
//
// Returns:
//   - T: The option value if valid, otherwise the default value
//
// Example:
//
//	// With default
//	timeout := GetOptDef(30*time.Second, opts...)
//
//	// With explicit option
//	opts := []time.Duration{60 * time.Second}
//	timeout := GetOptDef(30*time.Second, opts...) // returns 60s
//
//	// With zero value (uses default)
//	opts := []time.Duration{0}
//	timeout := GetOptDef(30*time.Second, opts...) // returns 30s
func GetOptDef[T comparable](def T, opts ...T) T {
	if len(opts) == 0 {
		return def
	}

	if len(opts) > 1 {
		panic(fmt.Sprintf("too many options: expected 0 or 1, got %d", len(opts)))
	}

	first := opts[0]
	if !validOpt(first) {
		return def
	}

	return first
}

// validOpt checks if an option is valid (non-zero and non-nil for pointers).
//
// Validity criteria:
//   - For non-pointer types: not equal to zero value
//   - For pointer types: not nil AND not pointing to zero value
//
// This function uses reflection to handle pointer types generically.
//
// Parameters:
//   - opt: The option value to validate
//
// Returns:
//   - bool: True if the option is valid, false otherwise
func validOpt[T comparable](opt T) bool {
	// Get the zero value for comparison
	var zero T

	// For pointer types, only check for nil (allow zero-value structs through)
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Pointer {
		// Pointer is valid if not nil
		return !v.IsNil()
	}

	// For non-pointer types, check if it's the zero value
	return opt != zero
}
