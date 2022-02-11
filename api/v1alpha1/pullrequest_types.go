/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sort"
	"time"

	"golang.org/x/oauth2"

	bitbucketClient "github.com/gfleury/go-bitbucket-v1"
	githubClient "github.com/google/go-github/v42/github"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PullRequestSpec defines the desired state of PullRequest
type PullRequestSpec struct {

	// GitProvider points at the object specifying the git provider, e.g. Bitbucket or Github
	// +kubebuilder:validation:Required
	GitProvider GitProvider `json:"gitProvider"`

	// TargetBranch points at the object specifying the target branch
	// +kubebuilder:validation:Required
	TargetBranch Branch `json:"targetBranch"`

	// Interval at which to reconcile the git provider.
	// +required
	Interval metav1.Duration `json:"interval"`
}

// PullRequestStatus defines the observed state of PullRequest
type PullRequestStatus struct {

	// The branches from which a pull requst was opened to the target branch
	SourceBranches []Branch `json:"sourceBranches,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PullRequest is the Schema for the pullrequests API
type PullRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PullRequestSpec   `json:"spec,omitempty"`
	Status PullRequestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PullRequestList contains a list of PullRequest
type PullRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PullRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PullRequest{}, &PullRequestList{})
}

func (pullRequest *PullRequest) GetBitbucketPullRequests(username string, password string) {
	basicAuth := bitbucketClient.BasicAuth{UserName: username, Password: password}
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Millisecond)
	ctx = context.WithValue(ctx, bitbucketClient.ContextBasicAuth, basicAuth)
	defer cancel()
	// TODO: remove for prod
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := bitbucketClient.NewAPIClient(
		ctx,
		bitbucketClient.NewConfiguration(pullRequest.Spec.GitProvider.Bitbucket.RestEndpoint),
	)
	//username := "admin"
	opts := map[string]interface{}{
		"direction": "INCOMING",
		"at":        pullRequest.Spec.TargetBranch.Name,
	}

	response, err := client.DefaultApi.GetPullRequestsPage(pullRequest.Spec.GitProvider.Bitbucket.Project, pullRequest.Spec.GitProvider.Bitbucket.Repository, opts)
	//response, err := client.DefaultApi.GetSSHKeys(username)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	prList, err := bitbucketClient.GetPullRequestsResponse(response)
	if err != nil {
		fmt.Println(err)
	}

	sourceBranches := make([]Branch, len(prList))
	for i := 0; i < len(prList); i++ {
		var tempBranch Branch
		tempBranch.Name = prList[i].FromRef.DisplayID
		tempBranch.Commit = prList[i].FromRef.LatestCommit
		sourceBranches[i] = tempBranch
	}

	pullRequest.Status.SourceBranches = sourceBranches

}

func (pullRequest *PullRequest) GetGithubPullRequests(accessToken string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := githubClient.NewClient(tc)

	opts := githubClient.PullRequestListOptions{Base: pullRequest.Spec.TargetBranch.Name}

	var prList []*githubClient.PullRequest
	var prResponse *githubClient.Response
	prList, prResponse, err := client.PullRequests.List(ctx, pullRequest.Spec.GitProvider.Github.Owner, pullRequest.Spec.GitProvider.Github.Repository, &opts)
	if err != nil {
		fmt.Println(prResponse)
		fmt.Println(err)
	}

	sourceBranches := make([]Branch, len(prList))
	for i := 0; i < len(prList); i++ {
		var tempBranch Branch
		tempBranch.Name = prList[i].GetHead().GetRef()
		tempBranch.SHA = prList[i].GetHead().GetSHA()
		sourceBranches[i] = tempBranch
	}

	pullRequest.Status.SourceBranches = sourceBranches

}

//GetLastCondition retruns the last condition based on the condition timestamp. if no condition is present it return false.
func (m *PullRequest) GetLastCondition() metav1.Condition {
	if len(m.Status.Conditions) == 0 {
		return metav1.Condition{}
	}
	//we need to make a copy of the slice
	copiedConditions := []metav1.Condition{}
	for _, condition := range m.Status.Conditions {
		ccondition := condition.DeepCopy()
		copiedConditions = append(copiedConditions, *ccondition)
	}
	sort.Slice(copiedConditions, func(i, j int) bool {
		return copiedConditions[i].LastTransitionTime.Before(&copiedConditions[j].LastTransitionTime)
	})
	return copiedConditions[len(copiedConditions)-1]
}

func (m *PullRequest) GetCondition(conditionType string) (metav1.Condition, bool) {
	for _, condition := range m.Status.Conditions {
		if condition.Type == conditionType {
			return condition, true
		}
	}
	return metav1.Condition{}, false
}

func (m *PullRequest) ReplaceCondition(c metav1.Condition) {
	if len(m.Status.Conditions) == 0 {
		m.Status.Conditions = append(m.Status.Conditions, c)
	} else {
		m.Status.Conditions[0] = c
	}
}

func (m *PullRequest) AddOrReplaceCondition(c metav1.Condition) {
	found := false
	for i, condition := range m.Status.Conditions {
		if c.Type == condition.Type {
			m.Status.Conditions[i] = c
			found = true
		}
	}
	if found == false {
		m.Status.Conditions = append(m.Status.Conditions, c)
	}
}
