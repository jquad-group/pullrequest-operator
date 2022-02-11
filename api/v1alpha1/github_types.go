package v1alpha1

type Github struct {

	// +kubebuilder:validation:Required
	Url string `json:"url"`

	// +kubebuilder:validation:Required
	Owner string `json:"owner"`

	// +kubebuilder:validation:Required
	Repository string `json:"repository"`
}
