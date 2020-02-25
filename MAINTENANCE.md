# Build

## Docker 

### Build Image

```bash
docker build -t tellusxdp/market-sdk-gateway:latest . && \
docker push tellusxdp/market-sdk-gateway:latest
```

## Release Binary

1. Build the binary and create the archive.

```bash
./build.sh
```

Upload `dist/tellus-market-sdk-gateway-linux-amd64.zip` to [github.com](https://github.com/tellusxdp/tellus-market-sdk-gateway/releases/tag/latest) .

2. Update latest tag.

```
git checkout master
git pull
git tag -d latest
git tag latest
git push -f origin latest
```
