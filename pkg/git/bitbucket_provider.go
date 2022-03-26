package v1alpha1

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"

	"net/http"
	"strings"
	"time"

	bitbucketClient "github.com/gfleury/go-bitbucket-v1"
	pullrequestv1alpha1 "github.com/jquad-group/pullrequest-operator/api/v1alpha1"
)

type BitbucketPoller struct {
}

func NewBitbucketPoller() *BitbucketPoller {
	return &BitbucketPoller{}
}

func (bitbucketPoller BitbucketPoller) Poll(accessToken string, pullRequest pullrequestv1alpha1.PullRequest) (pullrequestv1alpha1.Branches, error) {
	accessToken = strings.TrimSuffix(accessToken, "\n")
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Millisecond)
	ctx = context.WithValue(ctx, bitbucketClient.ContextAccessToken, accessToken)
	defer cancel()
	// TODO: remove for prod
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	bitbucketConfig := bitbucketClient.NewConfiguration(pullRequest.Spec.GitProvider.Bitbucket.RestEndpoint)
	client := bitbucketClient.NewAPIClient(
		ctx,
		bitbucketConfig,
	)

	opts := map[string]interface{}{
		"direction": "INCOMING",
		"at":        pullRequest.Spec.TargetBranch.Name,
	}

	var branches pullrequestv1alpha1.Branches

	response, err := client.DefaultApi.GetPullRequestsPage(pullRequest.Spec.GitProvider.Bitbucket.Project, pullRequest.Spec.GitProvider.Bitbucket.Repository, opts)
	fmt.Println(response)
	//response, err := client.DefaultApi.GetSSHKeys(username)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return branches, err
	}

	prList, err := bitbucketClient.GetPullRequestsResponse(response)
	if err != nil {
		fmt.Println(err)
		return branches, err
	}

	sourceBranches := make([]pullrequestv1alpha1.Branch, len(prList))
	for i := 0; i < len(prList); i++ {
		var tempBranch pullrequestv1alpha1.Branch
		tempBranch.Name = prList[i].FromRef.DisplayID
		tempBranch.Commit = prList[i].FromRef.LatestCommit
		pr, err := json.Marshal(prList[i])
		if err != nil {
			fmt.Println(err)
			return branches, nil
		}
		tempBranch.Details = string(pr)
		sourceBranches[i] = tempBranch
	}

	branches.Branches = sourceBranches

	return branches, nil

}
