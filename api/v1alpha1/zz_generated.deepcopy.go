// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Bitbucket) DeepCopyInto(out *Bitbucket) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Bitbucket.
func (in *Bitbucket) DeepCopy() *Bitbucket {
	if in == nil {
		return nil
	}
	out := new(Bitbucket)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Branch) DeepCopyInto(out *Branch) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Branch.
func (in *Branch) DeepCopy() *Branch {
	if in == nil {
		return nil
	}
	out := new(Branch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitProvider) DeepCopyInto(out *GitProvider) {
	*out = *in
	out.Bitbucket = in.Bitbucket
	out.Github = in.Github
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitProvider.
func (in *GitProvider) DeepCopy() *GitProvider {
	if in == nil {
		return nil
	}
	out := new(GitProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Github) DeepCopyInto(out *Github) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Github.
func (in *Github) DeepCopy() *Github {
	if in == nil {
		return nil
	}
	out := new(Github)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PullRequest) DeepCopyInto(out *PullRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PullRequest.
func (in *PullRequest) DeepCopy() *PullRequest {
	if in == nil {
		return nil
	}
	out := new(PullRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PullRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PullRequestList) DeepCopyInto(out *PullRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PullRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PullRequestList.
func (in *PullRequestList) DeepCopy() *PullRequestList {
	if in == nil {
		return nil
	}
	out := new(PullRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PullRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PullRequestSpec) DeepCopyInto(out *PullRequestSpec) {
	*out = *in
	out.GitProvider = in.GitProvider
	out.TargetBranch = in.TargetBranch
	out.Interval = in.Interval
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PullRequestSpec.
func (in *PullRequestSpec) DeepCopy() *PullRequestSpec {
	if in == nil {
		return nil
	}
	out := new(PullRequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PullRequestStatus) DeepCopyInto(out *PullRequestStatus) {
	*out = *in
	if in.SourceBranches != nil {
		in, out := &in.SourceBranches, &out.SourceBranches
		*out = make([]Branch, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PullRequestStatus.
func (in *PullRequestStatus) DeepCopy() *PullRequestStatus {
	if in == nil {
		return nil
	}
	out := new(PullRequestStatus)
	in.DeepCopyInto(out)
	return out
}
