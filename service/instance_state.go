package service

// InstanceState is interface that represents current instance state
type InstanceState interface {
	IsUp() bool
	IsFailed() bool
	IsMigrating() bool
	HasDiff() bool
}
