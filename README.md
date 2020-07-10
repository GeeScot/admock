# admock
AdMock: DNS service with DNS over HTTPS and network level ad-blocking.

See [codescot/blackhole](https://github.com/codescot/blackhole) for aggregated and sorted blacklist.

![](https://github.com/codescot/admock/workflows/Go/badge.svg)

# environment variables
```
# valid values: https, udp (note: using udp strategy may require elevated privileges, i.e. sudo)
# default: https
ADMOCK_STRATEGY=udp

# customise dns servers
# default: 1.1.1.1, 1.0.0.1
ADMOCK_DNS1=1.1.1.1
ADMOCK_DNS2=1.0.0.1 

# use config.json file for custom whitelist and blacklist
# default: n/a
ADMOCK_CONFIG=/full/path/to/config.json
```

# docker setup
```
docker build -t admock .
docker run \
 -d -p 53:53/udp \
 --restart=unless-stopped \
 --name admock admock
```
