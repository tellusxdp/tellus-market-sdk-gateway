# tellus-market-sdk-gateway

Market SDK Gateway

## 処理内容

JWTのバリデーションを行い、認証に通れば、別のサーバにリバースプロキシする

## 設定

config.ymlにて以下の項目を設定します

|                             | 意味                     | 例                                                   |
| --------------------------- | ----------------------- | --------------------------------------------------- |
| http.listen_address         | Listenするアドレス             | 127.0.0.1:8000                                      |
| http.tls.autocert.enabled   | 自動証明書発行を有効にするか      | true                               |
| http.tls.autocert.cache_dir | 証明書発行のキャッシュディレクトリ | /tmp/autocert                    |
| http.tls.certificate        | 証明書 | /opt/cert.pem    |
| http.tls.key                | 秘密鍵 | /opt/privkey.pem |
| upsteram.url                | 認証後に接続するサーバ             | https://www.example.com/         |
| upstream.headers            | プロキシ先に付与するリクエストヘッダ | {"Authorization": "Bearer token"} |
| private_key_url | JWTを検証する公開鍵をダウンロードするURL | https://sdk.tellusxdp.com/api/manager/v1/auth/public_keys |
| upstream        | 認証後に接続するサーバ             | https://www.example.com/                            |
| provider_name   | プロバイダ名                    | provider-a                                           |
| tool_id         | 商品ID                    | 1_9ffc0bb13148c605795b5bc22143b7b00c30ad            |
| tool_label      | 商品ラベル                 | product01                                           |



## Getting started

```bash
mkdir -p ~/go/src/github.com/tellusxdp
cd ~/go/src/github.com/tellusxdp
git clone https://github.com/tellusxdp/tellus-market-sdk-gateway
cd tellus-market-sdk-gateway
vi config.yml
go run main.go
```


## Docker

### Build Image

```bash
docker build -t tellusxdp/market-sdk-gateway:latest .
docker push tellusxdp/market-sdk-gateway:latest
```

### Using Image

```bash
docker run -it -p 8000:8000 -v `pwd`:/opt/market tellusxdp/market-sdk-gateway:latest
```