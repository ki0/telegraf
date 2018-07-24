# Devo Output Plugin
This plugin send metrics in influx data format to Devo platform by tcp, udp and tcp+tls connection.

### Configuration:

```toml
# Configuration for Devo platform to send metrics to
[[outputs.devo]]
  ## TCP endpoint for your devo entry point.
  # endpoint = "tcp://localhost:514"
  ## Prefix metrics name, syslog tag.
  # tag = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```
