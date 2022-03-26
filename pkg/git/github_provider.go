package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	githubClient "github.com/google/go-github/v42/github"
	pullrequestv1alpha1 "github.com/jquad-group/pullrequest-operator/api/v1alpha1"
	"golang.org/x/oauth2"
)

type GithubPoller struct {
}

func NewGithubPoller() *GithubPoller {
	return &GithubPoller{}
}

func (githubPoller GithubPoller) Poll(accessToken string, pullRequest pullrequestv1alpha1.PullRequest) (pullrequestv1alpha1.Branches, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := githubClient.NewClient(tc)

	opts := githubClient.PullRequestListOptions{Base: pullRequest.Spec.TargetBranch.Name}

	var branches pullrequestv1alpha1.Branches

	var prList []*githubClient.PullRequest
	var prResponse *githubClient.Response
	prList, prResponse, err := client.PullRequests.List(ctx, pullRequest.Spec.GitProvider.Github.Owner, pullRequest.Spec.GitProvider.Github.Repository, &opts)
	if err != nil {
		fmt.Println(prResponse)
		fmt.Println(err)
		return branches, err
	}

	sourceBranches := make([]pullrequestv1alpha1.Branch, len(prList))
	for i := 0; i < len(prList); i++ {
		var tempBranch pullrequestv1alpha1.Branch
		tempBranch.Name = prList[i].GetHead().GetRef()
		tempBranch.Commit = prList[i].GetHead().GetSHA()
		pr, err := json.Marshal(prList[i])
		if err != nil {
			fmt.Println(err)
			return branches, nil
		}
		tempBranch.Details = string(pr)
		sourceBranches[i] = tempBranch
	}
	branches.SetBranches(sourceBranches)

	return branches, nil
}
