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

package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	pipelinev1alpha1 "github.com/jquad-group/pullrequest-operator/api/v1alpha1"
	gitApi "github.com/jquad-group/pullrequest-operator/pkg/git"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	FIELD_MANAGER = "pullrequest-controller"

	BITBUCKET_PROVIDER_NAME = "Bitbucket"
	GITHUB_PROVIDER_NAME    = "Github"

	// Status
	ReconcileUnknown       = "Unknown"
	ReconcileError         = "Error"
	ReconcileErrorReason   = "Failed"
	ReconcileSuccess       = "Success"
	ReconcileSuccessReason = "Succeded"

	// Bitbucket and Github Secret Key
	SECRET_ACCESSTOKEN_KEY = "accessToken"
)

// PullRequestReconciler reconciles a PullRequest object
type PullRequestReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=pipeline.jquad.rocks,resources=pullrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pipeline.jquad.rocks,resources=pullrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pipeline.jquad.rocks,resources=pullrequests/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch;update;get;list;watch

func (r *PullRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := log.FromContext(ctx)

	var pullrequest pipelinev1alpha1.PullRequest
	if err := r.Get(ctx, req.NamespacedName, &pullrequest); err != nil {
		// return and dont requeue
		return ctrl.Result{}, nil
	}

	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   pipelinev1alpha1.GroupVersion.Group,
		Version: pipelinev1alpha1.GroupVersion.Version,
		Kind:    "PullRequest",
	})
	patch.SetNamespace(pullrequest.GetNamespace())
	patch.SetName(pullrequest.GetName())
	patchOptions := &client.PatchOptions{
		FieldManager: FIELD_MANAGER,
		Force:        pointer.Bool(true),
	}

	var prPoller gitApi.PullrequestPoller
	// Credentials for Github/Bitbucket are provided
	if len(pullrequest.Spec.GitProvider.SecretRef) > 0 {
		// try to find the provided secret on the cluster
		foundSecret := &v1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{Name: pullrequest.Spec.GitProvider.SecretRef, Namespace: pullrequest.Namespace}, foundSecret); err != nil {
			return r.ManageError(ctx, &pullrequest, req, err)
		}
		// validate the secret's format
		if err := Validate(&pullrequest, *foundSecret); err != nil {
			r.recorder.Event(&pullrequest, v1.EventTypeWarning, "Error", err.Error())
			return r.ManageError(ctx, &pullrequest, req, err)
		}
		prPoller = createGitPoller(&pullrequest, string(foundSecret.Data[SECRET_ACCESSTOKEN_KEY]))
	} else {
		prPoller = createGitPoller(&pullrequest, "")
	}

	newBranches, err := prPoller.Poll(pullrequest.Spec.TargetBranch.Name)
	if err != nil {
		r.recorder.Event(&pullrequest, v1.EventTypeWarning, "Error", err.Error())
		return r.ManageError(ctx, &pullrequest, req, err)
	}

	if !pullrequest.Status.SourceBranches.Equals(newBranches) {
		setDifferences := pullrequest.Status.SourceBranches.BranchSetDifference(newBranches)
		for i := 0; i < len(setDifferences); i++ {
			r.recorder.Event(&pullrequest, v1.EventTypeNormal, "Info", "New PR "+setDifferences[i].Name+"/"+setDifferences[i].Commit+" received.")
		}
		condition := metav1.Condition{
			Type:               ReconcileSuccess,
			LastTransitionTime: metav1.Now(),
			ObservedGeneration: pullrequest.GetGeneration(),
			Reason:             ReconcileSuccessReason,
			Status:             metav1.ConditionTrue,
			Message:            "Success",
		}
		pullrequest.AddOrReplaceCondition(condition)
		pullrequest.Status.SourceBranches.Branches = setDifferences
		patch.UnstructuredContent()["status"] = pullrequest.Status
		r.Status().Patch(ctx, patch, client.Apply, patchOptions)
	}

	return ctrl.Result{RequeueAfter: pullrequest.Spec.Interval.Duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PullRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("PullRequest")

	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinev1alpha1.PullRequest{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

func (r *PullRequestReconciler) ManageError(context context.Context, obj *pipelinev1alpha1.PullRequest, req ctrl.Request, message error) (reconcile.Result, error) {
	log := log.FromContext(context)
	if err := r.Get(context, types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, obj); err != nil {
		log.Error(err, "unable to get obj")
		return reconcile.Result{}, err
	}

	condition := metav1.Condition{
		Type:               ReconcileError,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: obj.GetGeneration(),
		Reason:             ReconcileErrorReason,
		Status:             metav1.ConditionFalse,
		Message:            message.Error(),
	}
	obj.AddOrReplaceCondition(condition)
	err := r.Status().Update(context, obj)
	if err != nil {
		log.Error(err, "unable to update status")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func Validate(pullrequest *pipelinev1alpha1.PullRequest, secret v1.Secret) error {
	if len(secret.Data[SECRET_ACCESSTOKEN_KEY]) <= 0 {
		return fmt.Errorf("invalid HTTP auth option: 'accessToken' must be set")
	}
	return nil
}

func createGitPoller(repo *pipelinev1alpha1.PullRequest, accessToken string) gitApi.PullrequestPoller {
	switch repo.Spec.GitProvider.Provider {
	case GITHUB_PROVIDER_NAME:
		return gitApi.NewGithubPoller(repo.Spec.GitProvider.Github.Url, accessToken, repo.Spec.GitProvider.InsecureSkipVerify, repo.Spec.GitProvider.Github.Owner, repo.Spec.GitProvider.Github.Repository)
	case BITBUCKET_PROVIDER_NAME:
		return gitApi.NewBitbucketPoller(repo.Spec.GitProvider.Bitbucket.RestEndpoint, accessToken, repo.Spec.GitProvider.InsecureSkipVerify, repo.Spec.GitProvider.Bitbucket.Project, repo.Spec.GitProvider.Bitbucket.Repository)
	}
	return nil
}
