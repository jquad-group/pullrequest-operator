package v1alpha1

type Branches struct {
	Branches []Branch `json:"branches,omitempty"`
}

func (branches *Branches) SetBranches(newBranches []Branch) {
	branches.Branches = newBranches
}

func (branches *Branches) GetBranches() []Branch {
	return branches.Branches
}

func (branches *Branches) Equals(newBranches Branches) bool {
	found := true

	if branches.GetSize() == 0 {
		return false
	}

	if branches.GetSize() != newBranches.GetSize() {
		return false
	}

	for i, branch := range branches.Branches {
		if !newBranches.Branches[i].Equals(branch) {
			found = false
			break
		}
	}

	return found

}

func (branches *Branches) GetSize() int {
	return len(branches.Branches)
}

func (branches *Branches) BranchSetDifference(newBranches Branches) (diff []Branch) {
	m := make(map[Branch]bool)

	for _, item := range branches.Branches {
		m[item] = true
	}

	for _, item := range newBranches.Branches {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}

	return
}
