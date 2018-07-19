package devo

import (
	"fmt"
	"net"
	"strings"

	syslog "github.com/RackSec/srslog"
	tlsint "github.com/influxdata/telegraf/internal/tls"
)

type Devo struct {
	Endpoint  string
	SyslogTag string
	conn      syslog
	tlsint.ClientConfig
	serializers.Serializer
}

var sampleConfig = `
  ## TCP endpoint for your devo entry point.
  endpoint = "localhost:514"
  ## Prefix metrics name, syslog tag.
  syslogtag = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (d *Devo) SampleConfig() string {
	return sampleConfig
}

func (d *Devo) Description() string {
	return "Configuration for Devo to send metrics to"
}

func (d *Devo) Connect() error {
	if len(d.Endpoint) == 0 {
		d.Servers = append(d.Servers, "tcp://127.0.0.1:514")
	}

	// Set tls config
	tlsConfig, err := d.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	spl := strings.SplitN(d.Endpoint, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid address: %s", sw.Address)
	}

	if d.SyslogTag == "" {
		d.SyslogTag = "test.keep.free:"
	}

	// Get Connections
	var c syslog
	if tlsCfg == nil {
		c, err = syslog.Dial(spl[0], spl[1], syslog.LOG_NOTICE, d.SyslogTag)
	} else {
		c, err = syslog.DialWithTLSConfig("tcp+tls", spl[1], syslog.LOG_NOTICE, d.SyslogTag, tlsConfig)
	}
	if err != nil {
		return err
	}

	d.conn = c
	return nil
}

func (d *Devo) Write(metrics []telegraf.Metric) error {
	if d.conn == nil {
		// previous write failed with permanent error and socket was closed.
		if err := d.Connect(); err != nil {
			return err
		}
	}

	for _, m := range metrics {
		bs, err := d.Serialize(m)
		if err != nil {
			//TODO log & keep going with remaining metrics
			return err
		}
		if _, err := d.conn.Write(bs); err != nil {
			//TODO log & keep going with remaining strings
			if err, ok := err.(net.Error); !ok || !err.Temporary() {
				// permanent error. close the connection
				d.Close()
				d.conn = nil
				return fmt.Errorf("closing connection: %v", err)
			}
			return err
		}
	}

	return nil
}

func (d *Devo) Close() error {
	// Closing all connections
	if d.conn == nil {
		return nil
	}
	err := d.conn.Close()
	d.conn = nil
	return nil
}

func (d *Devo) SetSerializer(s serializers.Serializer) {
	d.Serializer = s
}

func newDevo() *Devo {
	s, _ := serializers.NewInfluxSerializer()
	return &Devo{
		Serializer: s,
	}
}

func init() {
	outputs.Add("devo", func() telegraf.Output { return newDevo() })
}
