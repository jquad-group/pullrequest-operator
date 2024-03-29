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
	"sort"

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
	SourceBranches Branches `json:"sourceBranches,omitempty"`

	ETag string `json:"etag,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

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

// GetLastCondition retruns the last condition based on the condition timestamp. if no condition is present it return false.
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
