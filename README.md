# `cache-eks-creds`

Speed up your `kubectl` with EKS cluster

## What is this

Using EKS may slow down your `kubectl` command because not latency for communicating with clouds, but authentication mechanism.
If you use kubeconfig create by `aws eks update-kubeconfig` or `eksctl utils write-kubeconfig`, each command execution calls aws-cli and AWS API to get authentication token.
By author's easy bench, that call costs 3-5 seconds.

This command works as wrapper of aws-cli, and cache token to file.
If the cache not expired, return it to kubectl.
Of course, if the cache already expired, call AWS API and get new token.

> Above idea based on following article:
>
> - [[EKS] kubectlを高速化する - Qiita](https://qiita.com/masahata/items/e76ed2c91eeaa095d7c7)
>
> Thank you @buildsville for your helpful article.

## How to use

### By building from source

> We have no binary release yet.

Clone this repository.

Build by following command:

```sh
go build -o cache-eks-creds main.go
```

Copy of move binary to anywhere included to PATH.

Replace value from `aws` to `cache-eks-creds` of field `$.users[*].user.exec.command` in your `kubeconfig`.
Like following:

```yaml
users:
- name: user-name
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      args:
        - --region
        - ap-northeast-1
        - eks
        - get-token
        - --cluster-name
        - my-cluster
      #command: aws  # original
      command: cache-eks-creds
      env:
        - name: AWS_PROFILE
          value: my-profile
```
