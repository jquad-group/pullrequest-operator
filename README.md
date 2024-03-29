# Pull Request Operator 

The Pull Request operator checks a target branch in a repository for new pull requests at a specified interval. 

![Workflow](https://github.com/jquad-group/pullrequest-operator/blob/main/img/pullrequest-operator.svg)

# Installation 

Run the following command:

`kubectl apply -f https://github.com/jquad-group/pullrequest-operator/releases/latest/download/release.yaml` 

The operator is installed in the pullrequest-operator-system namespace.

After the installation of the operator, the PullRequest resource is added to the kubernetes cluster.

# Specification 

## Bitbucket

For the bitbucket provider a rest endpoint url must be specified, a project and the repository where the code resides. Currently only Bitbucket Server is supported.

```
apiVersion: pipeline.jquad.rocks/v1alpha1
kind: PullRequest
metadata:
  name: pullrequest-bitbucket-sample
spec:
  gitProvider:
    provider: Bitbucket
    secretRef: bitbucket-secret
    bitbucket:
      restEndpoint: https://bitbucket.jquad.rocks/rest
      project: jquad
      repository: microservice
  targetBranch: 
    name: refs/heads/main
  interval: 1m
```

## Github

For the github provider one must specifiy the url to the repository, the owner and the repository name. 

```
apiVersion: pipeline.jquad.rocks/v1alpha1
kind: PullRequest
metadata:
  name: pullrequest-bitbucket-sample
spec:
  gitProvider:
    provider: Github
    secretRef: github-secret
    github:
      url: https://github.com/rannox/microservice.git
      owner: rannox
      repository: microservice
  targetBranch: 
    name: refs/heads/main
  interval: 1m
```

# Authentication and Authorization

The Github and Bitbucket providers accept only an access token. 

## Bitbucket

In order to create an access token, go to `Profile->Account settings->HTTP access tokens->create token`. Encode the created token in base64 and save the value in a kubernetes `Secret` with the key `accessToken`:

```
apiVersion: v1
data:
  accessToken: BASE64
kind: Secret
metadata:
  name: bitbucket-secret
type: Opaque
```

# GitHub

```
apiVersion: v1
data:
  accessToken: BASE64 Personal Access Token
kind: Secret
metadata:
  name: github-secret
type: Opaque
```

# Status Example

The following status is created after a successful pull for the pull requests:

```
Status:
  Conditions:
    Last Transition Time:  2022-04-14T17:38:29Z
    Message:               Source branches reconciliation is successful.
    Observed Generation:   1
    Reason:                Succeded
    Status:                True
    Type:                  Success
  Source Branches:
    Branches:
      Commit:   e75d9b5beaf8dc12ac19ec0f72d254ad32edcc19
      Details:  {} # JSON representation of the response from Bitbucket or Github
      Name:     feature-kaniko
```
