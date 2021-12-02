个人自用。

```yaml
version: '3.7'

service:
  my-site-proxy:
    image: ghcr.io/trim21/my-site-proxy:latest
    restart: unless-stopped
    environment:
      RTORRENT_API_ENTRYPOINT: 'http://rtorrent.omv.trim21.me/RPC2'
      TRANSMISSION_API_ENTRYPOINT: "http://admin:password@192.168.1.3:8080"
    ports:
      - "3003:80"

```
