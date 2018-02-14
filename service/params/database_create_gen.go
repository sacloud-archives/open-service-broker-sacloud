package params

// DatabaseCreateParameter represents database-parameter
// for SAKURA Cloud Database Appliances
type DatabaseCreateParameter struct {
	SwitchID      int64    `json:"switch_id"`
	IPAddress     string   `json:"ipaddress"`
	MaskLen       int32    `json:"mask_len"`
	DefaultRoute  string   `json:"default_route"`
	UserName      string   `json:"user_name,omitempty"`
	Port          int32    `json:"portcomitempty"`
	BackupTime    string   `json:"backup_time,omitempty"`
	AllowNetworks []string `json:"allow_networks,omitempty"`
	PlanID        int
}
