package v1alpha1

type Bitbucket struct {

	// +kubebuilder:validation:Required
	RestEndpoint string `json:"restEndpoint"`

	// +kubebuilder:validation:Required
	Project string `json:"project"`

	// +kubebuilder:validation:Required
	Repository string `json:"repository"`
}
