package cmp

import (
	gocmp "github.com/google/go-cmp/cmp"
)

// CompareValue represents the value
// to be compared for use with the Equal function
type CompareValue struct {
	X interface{}
	Y interface{}
}

// Equal returns true if the given values ​​are equal
func Equal(values ...CompareValue) bool {
	for _, v := range values {
		res := gocmp.Equal(v.X, v.Y)
		if !res {
			return res
		}
	}
	return true
}
