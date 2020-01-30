# tellus-market-sdk-gateway

Market SDK Gateway

## 処理内容

JWTのバリデーションを行い、認証に通れば、別のサーバにリバースプロキシする

## 設定

config.ymlにて以下の項目を設定します

|                 | 意味                      | 例                                                   |
| --------------- | ----------------------- | --------------------------------------------------- |
| listen_address  | Listenするアドレス            | 127.0.0.1:8000                                      |
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