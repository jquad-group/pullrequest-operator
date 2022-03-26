package v1alpha1

import (
	pullrequestv1alpha1 "github.com/jquad-group/pullrequest-operator/api/v1alpha1"
)

type PullrequestPoller interface {
	Poll(accessToken string, pullRequest pullrequestv1alpha1.PullRequest) (pullrequestv1alpha1.Branches, error)
}
