# fastdns
FastDNS: fast as in quick to set up for DNS over HTTPS with network level ad-blocking

# environment variables
```
# defaults to https (valid values: https, udp)
# using udp strategy may require elevated privileges (sudo)
FASTDNS_STRATEGY=udp

# defaults to cloudflare dns if not specified
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
