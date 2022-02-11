package v1alpha1

type Branch struct {
	Name   string `json:"name"`
	SHA    string `json:"sha,omitempty"`
	Commit string `json:"commit,omitempty"`
}

func (currentBranch *Branch) equals(newBranch Branch) bool {
	if currentBranch.Name == newBranch.Name && currentBranch.SHA == newBranch.SHA && currentBranch.Commit == newBranch.Commit {
		return true
	} else {
		return false
	}
}
