package cohort

// LivenessChecker is any system used for keeping track of which systems are up
// and down
type LivenessChecker interface {
	// GetAliveHosts returns a list of hosts that are currently reachable
	GetAliveHosts() []string
}
