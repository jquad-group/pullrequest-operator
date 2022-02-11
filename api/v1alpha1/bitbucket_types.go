package v1alpha1

type Bitbucket struct {

	// +kubebuilder:validation:Required
	RestEndpoint string `json:"restEndpoint"`

	// +kubebuilder:validation:Required
	Project string `json:"project"`

	// +kubebuilder:validation:Required
	Repository string `json:"repository"`
}

/*
func Create(restEndpoint string, project string, repository string) *Bitbucket {
	return &Bitbucket{
		RestEndpoint: restEndpoint,
		Project:      project,
		Repository:   repository,
	}
}

func New() *Bitbucket {
	return &Bitbucket{
		RestEndpoint: "restEndpoint",
		Project:      "project",
		Repository:   "repository",
	}
}
*/
