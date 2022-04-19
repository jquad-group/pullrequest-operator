# Pull Request Operator 

The Pull Request operator checks a target branch in a repository for new pull requests at a specified interval. 

![Workflow](https://github.com/jquad-group/pullrequest-operator/blob/main/img/pullrequest-operator.svg)

# Specification 

## Bitbucket

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