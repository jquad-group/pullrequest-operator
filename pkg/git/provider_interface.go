package v1alpha1

import (
	pullrequestv1alpha1 "github.com/jquad-group/pullrequest-operator/api/v1alpha1"
)

type PullrequestPoller interface {
	Poll(branch string) (pullrequestv1alpha1.Branches, error)
}
