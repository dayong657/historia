package cohort

import "testing"

// we can't do much for permute since it's random, but we can make sure it
// doesn't crash

func TestPermute(t *testing.T) {
	permute([]string{})         // zero items
	permute([]string{"a"})      // one item
	permute([]string{"a", "b"}) // more than one item
}

type fakeLiveness struct {
	Hosts []string
}

func (this *fakeLiveness) GetAliveHosts() []string {
	return this.Hosts
}

func TestCohort(t *testing.T) {
	hosts := []string{"a", "b", "c", "d", "e"}
	me := 1
	mode := NewReadMajorityWriteMajority(len(hosts))
	live := fakeLiveness{hosts}
	c := NewCohort(me, hosts, mode, &live)
	c.GetAliveSet()

	nodes, err := c.getNodes(1)
	if err != nil || len(nodes) != 1 || nodes[0] != hosts[1] {
		t.Errorf("Expected to get us with one node, err: %s me: %s nodes: %s\n", err, hosts[1], nodes[0])
	}

	if hosts, err := c.GetCreateSet(); len(hosts) != mode.NodesNeededToCreate() || err != nil {
		t.Errorf("Bad create set")
	}

	if hosts, err := c.GetReadSet(); len(hosts) != mode.NodesNeededToRead() || err != nil {
		t.Errorf("Bad read set")
	}

	if hosts, err := c.GetUpdateSet(); len(hosts) != mode.NodesNeededToUpdate() || err != nil {
		t.Errorf("Bad update set")
	}

	if hosts, err := c.GetDeleteSet(); len(hosts) != mode.NodesNeededToDelete() || err != nil {
		t.Errorf("Bad delete set")
	}

	// make sure we can't get enough hosts now
	live.Hosts = []string{}

	if hosts, err := c.GetCreateSet(); hosts != nil || err == nil {
		t.Errorf("Bad create set, expected err and nil hosts")
	}

	if hosts, err := c.GetReadSet(); hosts != nil || err == nil {
		t.Errorf("Bad read set, expected err and nil hosts")
	}

	if hosts, err := c.GetUpdateSet(); hosts != nil || err == nil {
		t.Errorf("Bad update set, expected err and nil hosts")
	}

	if hosts, err := c.GetDeleteSet(); hosts != nil || err == nil {
		t.Errorf("Bad delete set, expected err and nil hosts")
	}

}
