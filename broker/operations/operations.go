package operations

const (
	// Provisioning represents the "provisioning" operation
	Provisioning = "provisioning"
	// Updating represents the "updating" operation
	Updating = "updating"
	// Deprovisioning represents the "deprovisioning" operation
	Deprovisioning = "deprovisioning"
	// Binding represents the "binding" operation
	Binding = "binding"
	// Unbinding represents the "unbinding" operation
	Unbinding = "unbinding"
	// StateInProgress represents the state of an operation that is still
	// pending completion
	StateInProgress = "in progress"
	// StateSucceeded represents the state of an operation that has
	// completed successfully
	StateSucceeded = "succeeded"
	// StateFailed represents the state of an operation that has
	// failed
	StateFailed = "failed"
	// StateGone is a pseudo oepration state represting the "state"
	// of an operation against an entity that no longer exists
	StateGone = "gone"
)
