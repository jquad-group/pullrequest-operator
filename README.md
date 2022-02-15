# Pull Request Operator 

# Specification 

# Bitbucket Auth

```
apiVersion: v1
data:
  # admin
  password: BASE64 
  # admin
  username: BASE64
kind: Secret
metadata:
  name: bitbucket-secret
type: Opaque
```

# Github Auth

```
apiVersion: v1
data:
  accessToken: BASE64 Personal Access Token
kind: Secret
metadata:
  name: github-secret
type: Opaque
```