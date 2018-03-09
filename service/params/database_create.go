package params

import (
	"fmt"

	"github.com/sacloud/open-service-broker-sacloud/util/validator"
)

// Validate performs parameter validation
func (p *DatabaseCreateParameter) Validate() error {

	required := map[string]interface{}{
		"switchID":     p.SwitchID,
		"ipaddress":    p.IPAddress,
		"maskLen":      p.MaskLen,
		"defaultRoute": p.DefaultRoute,
	}

	for k, v := range required {
		if !validator.Required(v) {
			return fmt.Errorf("%q is required", k)
		}
	}

	needIPv4 := map[string]string{
		"ipaddress":    p.IPAddress,
		"defaultRoute": p.DefaultRoute,
	}

	for k, v := range needIPv4 {
		if !validator.ValidIPv4Addr(v) {
			return fmt.Errorf("%q expects IPv4 format(xxx.xxx.xxx.xxx)", k)
		}
	}

	return nil
}
