package cohort

import (
	"errors"
	"math/rand"

	"github.com/josephlewis42/historia/checkup"
)

var (
	NotEnoughHostsError = errors.New("There are not enough alive hosts to complete the requests.")
)

// TCP ping and rmwm
func NewDefaultCohort(thishost int, hosts []string) Cohort {
	mode := NewReadMajorityWriteMajority(len(hosts))
	ckup := checkup.NewTCPCheckup(hosts)
	ckup.Start()
	return NewCohort(thishost, hosts, mode, ckup)
}

// Creates a new cohort
func NewCohort(thishost int, hosts []string, mode RWMode, ckup LivenessChecker) Cohort {
	c := Cohort{mode: mode, ckup: ckup, thishost: hosts[thishost]}
	return c
}

type Cohort struct {
	mode     RWMode
	ckup     LivenessChecker
	thishost string
}

func (this *Cohort) GetAliveSet() []string {
	return this.ckup.GetAliveHosts()
}

func (this *Cohort) GetCreateSet() ([]string, error) {
	numRequired := this.mode.NodesNeededToCreate()
	return this.getNodes(numRequired)
}

func (this *Cohort) GetReadSet() ([]string, error) {
	numRequired := this.mode.NodesNeededToRead()
	return this.getNodes(numRequired)
}

func (this *Cohort) GetUpdateSet() ([]string, error) {
	numRequired := this.mode.NodesNeededToUpdate()
	return this.getNodes(numRequired)
}

func (this *Cohort) GetDeleteSet() ([]string, error) {
	numRequired := this.mode.NodesNeededToDelete()
	return this.getNodes(numRequired)
}

func (this *Cohort) getNodes(num int) (nodes []string, err error) {
	if num == 1 {
		return []string{this.thishost}, nil
	}

	alive := this.ckup.GetAliveHosts()

	if len(alive) < num {
		return nil, NotEnoughHostsError
	}

	permute(alive) // we shuffle so we don't keep hitting the same hosts over and over

	alive = alive[:num] // get just the hosts we need

	// TODO make sure we're in the list for speed purposes
	/**for _, value := range alive {
		if value == thishost {
			return alive, nil
		}
	}

	alive[0] = thishost**/
	return alive[:num], nil
}

// permutes an array
func permute(hosts []string) {
	// http://stackoverflow.com/a/12267471
	for i := range hosts {
		j := rand.Intn(i + 1)
		hosts[i], hosts[j] = hosts[j], hosts[i]
	}
}
