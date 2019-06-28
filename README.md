# fastdns
FastDNS: fast as in quick to set up for DNS over HTTPS with network level ad-blocking

# docker setup
```
docker pull gurparit/fastdns
docker run \
 -d -v ~/path/to/config/fastdns/config.json:/app/config.json \
 -p 53:53/udp \
 --env FASTDNS_CONFIG=/app/config.json \
 --restart=unless-stopped
 --name fastdns gurparit/fastdns
```
