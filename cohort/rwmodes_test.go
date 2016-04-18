package cohort

import "testing"

func TestMajority(t *testing.T) {

	var data = []struct {
		Input    int
		Expected int
	}{
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 3},
		{5, 3},
		{7, 4},
		{1000, 501},
	}

	for _, tmp := range data {
		maj := majority(tmp.Input)

		if maj != tmp.Expected {
			t.Errorf("Got bad majority expected %d got %d\n", tmp.Expected, maj)
		}
	}
}

func TestROWA(t *testing.T) {
	var rowa = NewReadOneWriteAll(5)

	var data = []struct {
		Description string
		Expected    int
		Actual      int
	}{
		{"create", 5, rowa.NodesNeededToCreate()},
		{"read", 1, rowa.NodesNeededToRead()},
		{"update", 5, rowa.NodesNeededToUpdate()},
		{"delete", 5, rowa.NodesNeededToDelete()},
	}

	for tstno, tmp := range data {
		if tmp.Expected != tmp.Actual {
			t.Errorf("Got wrong number of hosts for %s (test #%d) expected: %d got: %d\n",
				tmp.Description, tstno, tmp.Expected, tmp.Actual)
		}
	}
}

func TestRMWM(t *testing.T) {
	var rowa = NewReadMajorityWriteMajority(5)

	var data = []struct {
		Description string
		Expected    int
		Actual      int
	}{
		{"create", 3, rowa.NodesNeededToCreate()},
		{"read", 3, rowa.NodesNeededToRead()},
		{"update", 5, rowa.NodesNeededToUpdate()},
		{"delete", 5, rowa.NodesNeededToDelete()},
	}

	for tstno, tmp := range data {
		if tmp.Expected != tmp.Actual {
			t.Errorf("Got wrong number of hosts for %s (test #%d) expected: %d got: %d\n",
				tmp.Description, tstno, tmp.Expected, tmp.Actual)
		}
	}
}
