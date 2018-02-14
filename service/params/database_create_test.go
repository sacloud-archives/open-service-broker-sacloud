package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseCreateParameterValidate(t *testing.T) {

	expects := []struct {
		name   string
		param  *DatabaseCreateParameter
		result bool
	}{
		{
			name: "SwitchID required",
			param: &DatabaseCreateParameter{
				SwitchID:     0,
				IPAddress:    "192.168.0.10",
				MaskLen:      24,
				DefaultRoute: "192.168.0.1",
			},
			result: false,
		},
		{
			name: "IPAddress required",
			param: &DatabaseCreateParameter{
				SwitchID:     999999999999,
				IPAddress:    "",
				MaskLen:      24,
				DefaultRoute: "192.168.0.1",
			},
			result: false,
		},
		{
			name: "IPAddress invalid format",
			param: &DatabaseCreateParameter{
				SwitchID:     999999999999,
				IPAddress:    "xxx.xxx.xxx.xxx",
				MaskLen:      24,
				DefaultRoute: "192.168.0.1",
			},
			result: false,
		},
		{
			name: "MaskLen required",
			param: &DatabaseCreateParameter{
				SwitchID:     999999999999,
				IPAddress:    "192.168.0.10",
				MaskLen:      0,
				DefaultRoute: "192.168.0.1",
			},
			result: false,
		},
		{
			name: "IPAddress required",
			param: &DatabaseCreateParameter{
				SwitchID:     999999999999,
				IPAddress:    "192.168.0.10",
				MaskLen:      24,
				DefaultRoute: "",
			},
			result: false,
		},
		{
			name: "IPAddress invalid format",
			param: &DatabaseCreateParameter{
				SwitchID:     999999999999,
				IPAddress:    "192.168.0.10",
				MaskLen:      24,
				DefaultRoute: "",
			},
			result: false,
		},
		{
			name: "Minimum valid params",
			param: &DatabaseCreateParameter{
				SwitchID:     999999999999,
				IPAddress:    "192.168.0.10",
				MaskLen:      24,
				DefaultRoute: "192.168.0.1",
			},
			result: true,
		},
	}

	for _, expect := range expects {
		p := expect.param
		err := p.Validate()
		t.Run(expect.name, func(t *testing.T) {
			assert.Equal(t, expect.result, err == nil)
		})
	}

}
