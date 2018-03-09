package params

// DatabaseCreateParameter represents database-parameter
// for SAKURA Cloud Database Appliances
type DatabaseCreateParameter struct {
	SwitchID      int64    `json:"switchID"`
	IPAddress     string   `json:"ipaddress"`
	MaskLen       int32    `json:"maskLen"`
	DefaultRoute  string   `json:"defaultRoute"`
	Username      string   `json:"username,omitempty"`
	Port          int32    `json:"port,omitempty"`
	BackupTime    string   `json:"backupTime,omitempty"`
	AllowNetworks []string `json:"allowNetworks,omitempty"`
	PlanID        int
}
