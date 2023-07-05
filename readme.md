A prometheus exporter for rtorrent, transmissionBT or qbittorrent.

# quick start

downloader-prom-exporter is release as docker image

## docker compose file

```yaml
version: "3.7"

service:
  downloader-prom-exporter:
    image: ghcr.io/trim21/downloader-prom-exporter:latest
    restart: unless-stopped
    environment:
      RTORRENT_API_ENTRYPOINT: "http://rtorrent.omv.trim21.me/RPC2"
      # or RTORRENT_API_ENTRYPOINT: "scgi://127.0.0.1:5000"
      # don't forget to mount socket file into container
      # or RTORRENT_API_ENTRYPOINT: "scgi:////home/ubuntu/.local/share/.rtorrent.sock"
      TRANSMISSION_API_ENTRYPOINT: "http://admin:password@192.168.1.3:8080"
      QBIT_API_ENTRYPOINT: "https://qb.omv.trim21.me"
    ports:
      - "8521:80"
```

## prometheus config

In your prometheus scrape config

```yaml
scrape_configs:
  - job_name: "downloader-exporter"
    static_configs:
      - targets: ["127.0.0.1:8521"]
```

# dashboard

https://grafana.com/grafana/dashboards/18986

# Transmission config details

If url path is empty string or `/`, download-exporter will use default rpc path `/transmission/rpc`.

You can also set non-standard transmission rpc path 

```shell
export TRANSMISSION_API_ENTRYPOINT="http://admin:password@127.0.0.1/tr/rpc"
```

# rTorrent config details

HTTP, TCP or unix socket protocol are all supported.


## unix domain socket

both relative path and absolute path are support

absolute file path

```shell
export RTORRENT_API_ENTRYPOINT="scgi:////home/ubuntu/.local/share/.rtorrent.sock"
```

relative file path, resolved from current working directory.

```shell
export RTORRENT_API_ENTRYPOINT="scgi:///.local/share/.rtorrent.sock"
```

(notice the triple slash `///` before file path)

exporter doesn't support user home expanding, do not use `~/...`.

also, don't forget to mount your unix socket file into docker container.

## TCP

```shell
export RTORRENT_API_ENTRYPOINT="scgi://127.0.0.1:5000"
```

## HTTP

if you are using apache, nginx or any http server to proxy scgi protocol to HTTP protocol:
```shell
export RTORRENT_API_ENTRYPOINT="http://rtorrent.omv.trim21.me/RPC2"
```

**you can't omit url path**

# Tips

In some case, for example, transmission's rpc will stop handling requests when moving files, rtorrent also may stop handling requests when in some case.

This will block the rpc request and may cause prometheus scraping failed with timeout.

If you want to avoid this, you will need to run dedicated exporter instance for each downloader,
so one downloader client won't affect other client's exporting.
