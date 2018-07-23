package devo

import (
	"fmt"
	"net"
	"strings"

	syslog "github.com/RackSec/srslog"
	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Devo struct {
	Endpoint  string
	SyslogTag string
	tlsint.ClientConfig
	serializers.Serializer
	SyslogWriter *syslog.Writer
}

var sampleConfig = `
  ## TCP endpoint for your devo entry point.
  endpoint = "tcp://localhost:514"
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
		d.Endpoint = "tcp://127.0.0.1:514"
	}

	// Set tls config
	tlsConfig, err := d.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	spl := strings.SplitN(d.Endpoint, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid address: %s", d.Endpoint)
	}

	if d.SyslogTag == "" {
		d.SyslogTag = "test.keep.free"
	}

	// Get Connections
	var w = &syslog.Writer{}
	if tlsConfig == nil {
		w, err = syslog.Dial(spl[0], spl[1], syslog.LOG_NOTICE, d.SyslogTag)
	} else {
		w, err = syslog.DialWithTLSConfig("tcp+tls", spl[1], syslog.LOG_NOTICE, d.SyslogTag, tlsConfig)
	}
	if err != nil {
		return err
	}

	d.SyslogWriter = w
	return nil
}

func (d *Devo) Write(metrics []telegraf.Metric) error {
	if d.SyslogWriter == nil {
		// previous write failed with permanent error and socket was closed.
		if err := d.Connect(); err != nil {
			return err
		}
	}

	for _, m := range metrics {
		bs, err := d.Serialize(m)
		if err != nil {
			return err
		}
		if _, err := d.SyslogWriter.Write(bs); err != nil {
			if err, ok := err.(net.Error); !ok || !err.Temporary() {
				d.Close()
				d.SyslogWriter = nil
				return fmt.Errorf("closing connection: %v", err)
			}
			return err
		}
	}

	return nil
}

func (d *Devo) Close() error {
	// Closing all connections
	if d.SyslogWriter == nil {
		return nil
	}
	err := d.SyslogWriter.Close()
	d.SyslogWriter = nil
	return err
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
