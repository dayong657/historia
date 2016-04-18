package cohort

import "math"

type ModeCount func(totalNumberOfNodes int) (numberRequired int)

func majority(totalNumberOfNodes int) (numberRequired int) {
	return int(math.Floor(float64(totalNumberOfNodes)/2.0) + 1)
}

type RWMode interface {
	NodesNeededToCreate() int
	NodesNeededToUpdate() int
	NodesNeededToDelete() int
	NodesNeededToRead() int
}

func NewReadOneWriteAll(numberOfNodes int) RWMode {
	return readOneWriteAll{numberOfNodes}
}

type readOneWriteAll struct {
	numberOfNodes int
}

func (r readOneWriteAll) NodesNeededToCreate() int {
	return r.numberOfNodes
}

func (r readOneWriteAll) NodesNeededToRead() int {
	return 1
}

func (r readOneWriteAll) NodesNeededToUpdate() int {
	return r.numberOfNodes
}

func (r readOneWriteAll) NodesNeededToDelete() int {
	return r.numberOfNodes
}

func NewReadMajorityWriteMajority(numberOfNodes int) RWMode {
	return rmwm{numberOfNodes}
}

type rmwm struct {
	numberOfNodes int
}

func (r rmwm) NodesNeededToCreate() int {
	return majority(r.numberOfNodes)
}

func (r rmwm) NodesNeededToRead() int {
	return majority(r.numberOfNodes)
}

func (r rmwm) NodesNeededToUpdate() int {
	return r.numberOfNodes
}

func (r rmwm) NodesNeededToDelete() int {
	return r.numberOfNodes
}
