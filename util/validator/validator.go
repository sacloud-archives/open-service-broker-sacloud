package validator

import (
	"net"
	"reflect"
	"strings"
	"time"
)

var numericZeros = []interface{}{
	int(0),
	int8(0),
	int16(0),
	int32(0),
	int64(0),
	uint(0),
	uint8(0),
	uint16(0),
	uint32(0),
	uint64(0),
	float32(0),
	float64(0),
}

// IsEmpty is copied from github.com/stretchr/testify/assert/assetions.go
func IsEmpty(object interface{}) bool {

	if object == nil {
		return true
	} else if object == "" {
		return true
	} else if object == false {
		return true
	}

	for _, v := range numericZeros {
		if object == v {
			return true
		}
	}

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
	case reflect.Map:
		fallthrough
	case reflect.Slice, reflect.Chan:
		{
			return (objValue.Len() == 0)
		}
	case reflect.Struct:
		switch object.(type) {
		case time.Time:
			return object.(time.Time).IsZero()
		}
	case reflect.Ptr:
		{
			if objValue.IsNil() {
				return true
			}
			switch object.(type) {
			case *time.Time:
				return object.(*time.Time).IsZero()
			default:
				return false
			}
		}
	}
	return false
}

// Required validates that value is not empty
func Required(v interface{}) bool {
	return !IsEmpty(v)
}

// ValidIPv4Addr validates that value is valid ipv4 address
func ValidIPv4Addr(addr string) bool {
	// if target is empty, return OK(Use Required if necessary)
	if addr == "" {
		return true
	}

	ip := net.ParseIP(addr)
	if ip == nil || !strings.Contains(addr, ".") {
		return false
	}
	return true
}
