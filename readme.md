A prometheus exporter for rtorrent, transmissionBT or qbittorrent.

```yaml
version: "3.7"

service:
  my-site-proxy:
    image: ghcr.io/trim21/my-site-proxy:latest
    restart: unless-stopped
    environment:
      RTORRENT_API_ENTRYPOINT: "http://rtorrent.omv.trim21.me/RPC2"
      # RTORRENT_API_ENTRYPOINT: "scgi://127.0.0.1:5000"
      TRANSMISSION_API_ENTRYPOINT: "http://admin:password@192.168.1.3:8080"
      QBIT_API_ENTRYPOINT: "https://qb.omv.trim21.me"
    ports:
      - "80:80"
```

```yaml
scrape_configs:
  - job_name: "transmission"
    metrics_path: /transmission/metrics
    static_configs:
      - targets: ["my-site-proxy:80"]
  - job_name: "rtorrent"
    metrics_path: /metrics # this path also export go runtime metrics
    static_configs:
      - targets: ["my-site-proxy:80"]
  - job_name: "qbittorrent"
    metrics_path: /qbit/metrics
    static_configs:
      - targets: ["my-site-proxy:80"]
```
