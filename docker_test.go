package dockertest

import (
	"net"
	"strconv"
	"testing"
	"time"

	"os"
	"regexp"

	"github.com/stretchr/testify/require"
)

func TestParsePorts(t *testing.T) {

	c := Container{}

	c.parsePorts("")
	require.NotNil(t, c.ports)
	require.Len(t, c.ports, 0)

	c.parsePorts("6379/tcp -> 0.0.0.0:32815")
	require.NotNil(t, c.ports)
	require.Len(t, c.ports, 1)
	require.Equal(t, 32815, c.ports[6379])

	c.parsePorts("6379/tcp -> 0.0.0.0:32815\n6380/udp -> 0.0.0.0:32816")
	require.NotNil(t, c.ports)
	require.Len(t, c.ports, 2)
	require.Equal(t, 32815, c.ports[6379])
	require.Equal(t, 32816, c.ports[6380])
}

func TestRun(t *testing.T) {

	con := Run("redis")
	defer con.Close()

	ip := "127.0.0.1"
	// for docker-machine
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		ipRegex := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
		ip = ipRegex.FindString(host)
	}

	require.Equal(t, ip, con.Host())

	port := con.ports[6379]

	require.Equal(t, net.JoinHostPort(ip, strconv.Itoa(port)), con.Addr(6379))
	require.Equal(t, port, con.Port(6379))

}

func TestWaitPort(t *testing.T) {

	con := Run("redis")
	defer con.Close()

	p := con.WaitPort(6379, 1*time.Second)
	require.NotZero(t, p)

}

func TestWaitHTTP(t *testing.T) {

	con := Run("nginx")
	defer con.Close()

	p := con.WaitHTTP(80, "/", 1*time.Second)
	require.NotZero(t, p)

}

func TestParseIp(t *testing.T) {
	c := Container{}

	c.parseIp("127.0.0.1")
	require.NotNil(t, c.host)
	require.Equal(t, "127.0.0.1", c.host)

	c.parseIp("tcp://192.168.99.100:2376")
	require.NotNil(t, c.host)
	require.Equal(t, "192.168.99.100", c.host)
}
