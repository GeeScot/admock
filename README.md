# fastdns
FastDNS: fast as in quick to set up for DNS over HTTPS with network level ad-blocking

# docker setup
```
docker build -t fastdns .
docker run \
 -d -p 53:53/udp \
 --restart=unless-stopped \
 --name fastdns fastdns
```
