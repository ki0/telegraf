package devo

import (
	"bufio"
	"bytes"
	"net"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevo_tcp(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	d := newDevo()
	d.Endpoint = "tcp://" + listener.Addr().String()
	d.SyslogTag = "test.keep.free"

	err = d.Connect()
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)

	testDevo_stream(t, d, lconn)
}

func TestDevo_udp(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	d := newDevo()
	d.Endpoint = "udp://" + listener.LocalAddr().String()
	d.SyslogTag = "test.keep.free"

	err = d.Connect()
	require.NoError(t, err)

	testDevo_packet(t, d, listener)
}

func testDevo_stream(t *testing.T, d *Devo, lconn net.Conn) {
	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := d.Serialize(metrics[0])
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := d.Serialize(metrics[1])

	err := d.Write(metrics)
	require.NoError(t, err)

	scnr := bufio.NewScanner(lconn)
	require.True(t, scnr.Scan())
	mstr1in := scnr.Text() + "\n"
	require.True(t, scnr.Scan())
	mstr2in := scnr.Text() + "\n"

	assert.Contains(t, mstr1in, string(mbs1out))
	assert.Contains(t, mstr2in, string(mbs2out))
}

func testDevo_packet(t *testing.T, d *Devo, lconn net.PacketConn) {
	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := d.Serialize(metrics[0])
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := d.Serialize(metrics[1])

	err := d.Write(metrics)
	require.NoError(t, err)

	buf := make([]byte, 256)
	var mstrins []string
	for len(mstrins) < 2 {
		n, _, err := lconn.ReadFrom(buf)
		require.NoError(t, err)
		for _, bs := range bytes.Split(buf[:n], []byte{'\n'}) {
			if len(bs) == 0 {
				continue
			}
			mstrins = append(mstrins, string(bs)+"\n")
		}
	}
	require.Len(t, mstrins, 2)

	assert.Contains(t, mstrins[0], string(mbs1out))
	assert.Contains(t, mstrins[1], string(mbs2out))
}
