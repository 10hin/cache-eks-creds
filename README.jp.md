# `cache-eks-creds`

EKSでの `kubectl` を高速化する

## これは何

EKSを使うと `kubectl` コマンドが遅くなりますが、それはクラウドと通信する遅延だけでなく、認証の仕組みに原因があります。
もしあなたが `aws eks update-kubeconfig` や `eksctl utils write-kubeconfig` で作成した `kubeconfig` を利用しているなら、各コマンド実行ごとに、認証トークンを取得するためaws-cliを通じてAWS APIが呼び出されています。
作者のが簡単に計測したところではこの呼び出しが3-5秒程度かかっています。

このコマンドはaws-cliのラッパーとして機能して、取得したトークンをファイルにキャッシュします。
もしキャッシュが期限内であれば、 `kubectl` にキャッシュされたトークンを返します。
もちろん、すでにキャッシュが期限切れしていれば、AWS APIを呼び出して新しいトークンを取得します。

> 上記のアイディアは以下の記事を参考にさせていただきました
>
> - [[EKS] kubectlを高速化する - Qiita](https://qiita.com/masahata/items/e76ed2c91eeaa095d7c7)
>
> @buildsville さんに、ここでお礼を申し上げます。有益な記事をありがとうございます。

## 使い方

### ソースコードからビルドする

> まだバイナリリリースはありません。

このリポジトリをcloneします。

次のコマンドでビルドします:

```sh
go build -o cache-eks-creds main.go
```

PATHの通った場所にバイナリをコピー/移動します。

`kubeconfig` の `$.users[*].user.exec.command` フィールドの値を `aws` から `cache-eks-creds` に置き換えます。

例:

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
