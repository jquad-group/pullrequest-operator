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
	Endpoint           string
	AccessToken        string
	InsecureSkipVerify bool
	Project            string
	Repository         string
}

func NewBitbucketPoller(endpoint string, accessToken string, insecureSkipVerify bool, project string, repository string) *BitbucketPoller {
	return &BitbucketPoller{
		Endpoint:           endpoint,
		AccessToken:        accessToken,
		InsecureSkipVerify: insecureSkipVerify,
		Project:            project,
		Repository:         repository,
	}
}

func (bitbucketPoller BitbucketPoller) Poll(branch string, etag string) (pullrequestv1alpha1.Branches, string, error) {
	accessToken := strings.TrimSuffix(bitbucketPoller.AccessToken, "\n")
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Millisecond)
	if len(bitbucketPoller.AccessToken) > 0 {
		ctx = context.WithValue(ctx, bitbucketClient.ContextAccessToken, accessToken)
	}
	defer cancel()
	if bitbucketPoller.InsecureSkipVerify {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	bitbucketConfig := bitbucketClient.NewConfiguration(bitbucketPoller.Endpoint)
	client := bitbucketClient.NewAPIClient(
		ctx,
		bitbucketConfig,
	)

	opts := map[string]interface{}{
		"direction": "INCOMING",
		"at":        branch,
	}

	var branches pullrequestv1alpha1.Branches

	response, err := client.DefaultApi.GetPullRequestsPage(bitbucketPoller.Project, bitbucketPoller.Repository, opts)
	fmt.Println(response)
	//response, err := client.DefaultApi.GetSSHKeys(username)
	if err != nil {
		//fmt.Printf("%s\n", err.Error())
		return branches, "", err
	}

	prList, err := bitbucketClient.GetPullRequestsResponse(response)
	if err != nil {
		//fmt.Println(err)
		return branches, "", err
	}

	sourceBranches := make([]pullrequestv1alpha1.Branch, len(prList))
	for i := 0; i < len(prList); i++ {
		var tempBranch pullrequestv1alpha1.Branch
		tempBranch.Name = prList[i].FromRef.DisplayID
		tempBranch.Commit = prList[i].FromRef.LatestCommit
		pr, err := json.Marshal(prList[i])
		if err != nil {
			//fmt.Println(err)
			return branches, "", err
		}
		tempBranch.Details = string(pr)
		sourceBranches[i] = tempBranch
	}

	branches.Branches = sourceBranches

	return branches, "", nil

}
