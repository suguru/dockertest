package dockertest

import (
	"bytes"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	portRegex = regexp.MustCompile(`([0-9]+)\/(.+?)\s\->.+?:([0-9]+)`)
)

// Container is docker container instance.
type Container struct {
	containerID string
	image       string
	ports       map[int]int
	networks    map[int]string
	host        string
}

// Run image and returns docker container.
func Run(image string, args ...string) *Container {

	// run and get containerID
	containerID, err := run("docker", append([]string{"run", "-P", "-d", image}, args...)...)
	if err != nil {
		log.Fatalf("failed run docker image:%s args:%v", image, args)
	}

	// get port map
	ports, err := run("docker", "port", containerID)
	if err != nil {
		log.Fatalf("failed get ports image:%s", image)
	}

	host := "127.0.0.1"

	// for docker-machine
	if os.Getenv("DOCKER_HOST") != "" {
		host = os.Getenv("DOCKER_HOST")
	}

	c := &Container{
		containerID: containerID,
		image:       image,
		host:        host,
	}
	c.parsePorts(ports)
	return c
}

// Close docker containerg.
func (c *Container) Close() {
	// ignore errors on stop, wait and remove
	run("docker", "stop", c.containerID)
	// wait until docker stops
	run("docker", "wait", c.containerID)
	// remove the container
	run("docker", "rm", c.containerID)
}

// Host returns host IP which runs docker.
func (c *Container) Host() string {
	return c.host
}

// WaitPort waits until port available.
func (c *Container) WaitPort(port int, timeout time.Duration) int {
	// wait until port available
	p := c.ports[port]
	if p == 0 {
		log.Fatalf("port %d is not exposed on %s", port, c.image)
	}

	nw := c.networks[port]
	if nw == "" {
		log.Fatalf("network not described on %s", c.image)
	}

	_, err := net.DialTimeout(nw, c.Addr(port), timeout)
	if err != nil {
		log.Fatalf("port not available for %f seconds", timeout.Seconds())
	}
	return p
}

// WaitHTTP waits until http available
func (c *Container) WaitHTTP(port int, path string, timeout time.Duration) int {
	p := c.ports[port]
	if p == 0 {
		log.Fatalf("port %d is not exposed on %s", port, c.image)
	}
	now := time.Now()
	end := now.Add(timeout)
	for {
		cli := http.Client{Timeout: timeout}
		res, err := cli.Get("http://" + c.Addr(port) + path)
		if err != nil {
			if time.Now().After(end) {
				log.Fatalf("http not available on port %d for %s err:%v", port, c.image, err)
			}
			// sleep 1 sec to retry
			time.Sleep(1 * time.Second)
			continue
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			if time.Now().After(end) {
				log.Fatalf("http has not valid status code on port %d for %s code:%d", port, c.image, res.StatusCode)
			}
			// sleep 1 sec to retry
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	return p
}

// Port returns exposed port in docker host.
func (c *Container) Port(port int) int {
	return c.ports[port]
}

// Addr returns exposed address like 127.0.0.1:6379.
func (c *Container) Addr(port int) string {
	exposed := c.Port(port)
	return net.JoinHostPort(c.host, strconv.Itoa(exposed))
}

// run command and get result.
func run(name string, args ...string) (out string, err error) {

	cmd := exec.Command(name, args...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err = cmd.Run(); err != nil {
		return
	}

	if cmd.ProcessState.Success() {
		return strings.TrimSpace(stdout.String()), nil
	}

	err = errors.New("command execution failed " + stderr.String())
	return
}

func (c *Container) parsePorts(lines string) {

	matches := portRegex.FindAllStringSubmatch(lines, -1)
	c.ports = make(map[int]int, len(matches))
	c.networks = make(map[int]string, len(matches))

	for _, match := range matches {
		p1, _ := strconv.Atoi(match[1])
		p2, _ := strconv.Atoi(match[3])
		c.ports[p1] = p2
		c.networks[p1] = match[2]
	}

}
