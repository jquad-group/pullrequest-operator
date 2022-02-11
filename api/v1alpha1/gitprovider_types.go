package v1alpha1

const (
	BITBUCKET_PROVIDER_NAME = "Bitbucket"
	GITHUB_PROVIDER_NAME    = "Github"
)

type GitProvider struct {

	// Git Provider type
	// +kubebuilder:validation:Enum=Bitbucket;Github
	// +kubebuilder:validation:Required
	Provider string `json:"provider"`

	// Git Provider credentials
	// +kubebuilder:validation:Optional
	SecretRef string `json:"secretRef"`

	// +kubebuilder:validation:Optional
	Bitbucket Bitbucket `json:"bitbucket"`

	// +kubebuilder:validation:Optional
	Github Github `json:"github"`
}
