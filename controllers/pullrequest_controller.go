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
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	pipelinev1alpha1 "github.com/jquad-group/pullrequest-operator/api/v1alpha1"
)

const (
	BITBUCKET_PROVIDER_NAME = "Bitbucket"
	GITHUB_PROVIDER_NAME    = "Github"

	// Status
	ReconcileUnknown       = "Unknown"
	ReconcileError         = "Error"
	ReconcileErrorReason   = "Failed"
	ReconcileSuccess       = "Success"
	ReconcileSuccessReason = "Succeded"

	// Bitbucket Secret keys
	BITBUCKET_SECRET_USERNAME_KEY = "username"
	BITBUCKET_SECRET_PASSWORD_KEY = "password"

	// Github Secret keys
	GITHUB_SECRET_ACCESSTOKEN_KEY = "accessToken"
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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PullRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *PullRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := log.FromContext(ctx)

	var pullrequest pipelinev1alpha1.PullRequest
	if err := r.Get(ctx, req.NamespacedName, &pullrequest); err != nil {
		// return and dont requeue
		return ctrl.Result{}, nil
	}

	foundSecret := &v1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: pullrequest.Spec.GitProvider.SecretRef, Namespace: pullrequest.Namespace}, foundSecret); err != nil {
		return r.ManageError(ctx, &pullrequest, req, err)
	}

	if err := Validate(&pullrequest, *foundSecret); err != nil {
		r.recorder.Event(&pullrequest, v1.EventTypeWarning, "Error", err.Error())
		return r.ManageError(ctx, &pullrequest, req, err)
	}

	if pullrequest.Spec.GitProvider.Provider == BITBUCKET_PROVIDER_NAME {
		newBranches, err := pullrequest.GetBitbucketPullRequests(string(foundSecret.Data[BITBUCKET_SECRET_USERNAME_KEY]), string(foundSecret.Data[BITBUCKET_SECRET_PASSWORD_KEY]))
		if err != nil {
			r.recorder.Event(&pullrequest, v1.EventTypeWarning, "Error", err.Error())
			return r.ManageError(ctx, &pullrequest, req, err)
		}
		if !pullrequest.Status.SourceBranches.Equals(newBranches) {
			setDifferences := pullrequest.Status.SourceBranches.BranchSetDifference(newBranches)
			for i := 0; i < len(setDifferences); i++ {
				r.recorder.Event(&pullrequest, v1.EventTypeNormal, "Info", "New PR "+setDifferences[i].Name+"/"+setDifferences[i].Commit+" received.")
			}
			return r.SourceBranchIsUpdatedStatus(ctx, &pullrequest, req, newBranches, "Source branches reconciliation is successful.")
		}
	}

	if pullrequest.Spec.GitProvider.Provider == GITHUB_PROVIDER_NAME {
		newBranches, err := pullrequest.GetGithubPullRequests(string(foundSecret.Data[GITHUB_SECRET_ACCESSTOKEN_KEY]))
		if err != nil {
			r.recorder.Event(&pullrequest, v1.EventTypeWarning, "Error", err.Error())
			return r.ManageError(ctx, &pullrequest, req, err)
		}
		if !pullrequest.Status.SourceBranches.Equals(newBranches) {
			setDifferences := pullrequest.Status.SourceBranches.BranchSetDifference(newBranches)
			for i := 0; i < len(setDifferences); i++ {
				r.recorder.Event(&pullrequest, v1.EventTypeNormal, "Info", "New PR "+setDifferences[i].Name+"/"+setDifferences[i].Commit+" received.")
			}
			return r.SourceBranchIsUpdatedStatus(ctx, &pullrequest, req, newBranches, "Source branches reconciliation is successful.")
		}
	}

	return ctrl.Result{RequeueAfter: pullrequest.Spec.Interval.Duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PullRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("PullRequest")

	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinev1alpha1.PullRequest{}).
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

func (r *PullRequestReconciler) SourceBranchIsUpdatedStatus(context context.Context, obj *pipelinev1alpha1.PullRequest, req ctrl.Request, newBranches pipelinev1alpha1.Branches, message string) (reconcile.Result, error) {
	log := log.FromContext(context)
	if err := r.Get(context, types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, obj); err != nil {
		log.Error(err, "unable to get obj")
		return reconcile.Result{}, err
	}

	obj.Status.SourceBranches = newBranches
	condition := metav1.Condition{
		Type:               ReconcileSuccess,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: obj.GetGeneration(),
		Reason:             ReconcileSuccessReason,
		Status:             metav1.ConditionTrue,
		Message:            message,
	}
	obj.AddOrReplaceCondition(condition)
	err := r.Status().Update(context, obj)
	if err != nil {
		log.Error(err, "unable to update status")
		return reconcile.Result{}, err
	}
	return reconcile.Result{RequeueAfter: obj.Spec.Interval.Duration}, nil
}

func (r *PullRequestReconciler) SourceBranchIsNotUpdatedStatus(context context.Context, obj *pipelinev1alpha1.PullRequest, req ctrl.Request, message string) (reconcile.Result, error) {
	log := log.FromContext(context)
	if err := r.Get(context, types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, obj); err != nil {
		log.Error(err, "unable to get obj")
		return reconcile.Result{}, err
	}

	condition := metav1.Condition{
		Type:               ReconcileSuccess,
		ObservedGeneration: obj.GetGeneration(),
		Reason:             ReconcileSuccessReason,
		Status:             metav1.ConditionTrue,
		Message:            message,
	}
	obj.AddOrReplaceCondition(condition)
	err := r.Status().Update(context, obj)
	if err != nil {
		log.Error(err, "unable to update status")
		return reconcile.Result{}, err
	}
	return reconcile.Result{RequeueAfter: obj.Spec.Interval.Duration}, nil
}

func Validate(pullrequest *pipelinev1alpha1.PullRequest, secret v1.Secret) error {
	switch pullrequest.Spec.GitProvider.Provider {
	case BITBUCKET_PROVIDER_NAME:
		if len(secret.Data[BITBUCKET_SECRET_USERNAME_KEY]) <= 0 && len(secret.Data[BITBUCKET_SECRET_USERNAME_KEY]) <= 0 {
			return fmt.Errorf("invalid HTTP auth option: 'password' requires 'username' to be set")
		}
	case GITHUB_PROVIDER_NAME:
		if len(secret.Data[GITHUB_SECRET_ACCESSTOKEN_KEY]) <= 0 {
			return fmt.Errorf("invalid HTTP auth option: 'accessToken' must be set")
		}

	}
	return nil
}
