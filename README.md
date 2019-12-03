# fastdns
FastDNS: fast as in quick to set up for DNS over HTTPS with network level ad-blocking

# dns server environment variables (with cloudflare defaults)
```
FASTDNS_DNS1=1.1.1.1 
FASTDNS_DNS2=1.0.0.1 
```

# docker setup
```
docker build -t fastdns .
docker run \
 -d -p 53:53/udp \
 --restart=unless-stopped \
 --name fastdns fastdns
```
