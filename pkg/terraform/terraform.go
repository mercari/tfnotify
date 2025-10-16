package terraform

const (
	// ExitPass is status code zero
	ExitPass int = iota

	// ExitFail is status code non-zero
	ExitFail
)
