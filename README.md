# tellus-market-sdk-gateway

Market SDK Gateway

## 処理内容

JWTのバリデーションを行い、認証に通れば、別のサーバにリバースプロキシする

## 設定

config.ymlにて以下の項目を設定します

|                 | 意味                      | 例                                                   |
| --------------- | ----------------------- | --------------------------------------------------- |
| listen_address  | Listenするアドレス            | 127.0.0.1:8000                                      |
| private_key_url | JWTを検証する公開鍵をダウンロードするURL | https://www.tellusxdp.com/market/api/v1/public_keys |
| upstream        | 認証後に接続するサーバ             | https://www.example.com/                            |
| product         | 商品ID                    | product01                                           |



## Getting started

```bash
mkdir -p ~/go/src/github.com/tellusxdp
cd ~/go/src/github.com/tellusxdp
git clone https://github.com/tellusxdp/tellus-market-sdk-gateway
cd tellus-market-sdk-gateway
vi config.yml
go run main.go
```
