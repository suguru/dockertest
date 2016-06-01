package dockertest

import (
	"bytes"
	"log"
	"net"
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
	containerID := run("docker", append([]string{"run", "-P", "-d", image}, args...)...)

	// get port map
	ports := run("docker", "port", containerID)

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

func (c *Container) waitPort(port int) int {
	// wait until port available
	p := c.ports[port]
	if p == 0 {
		log.Fatalf("port %d is not exposed on %s", port, c.image)
	}

	nw := c.networks[port]
	if nw == "" {
		log.Fatalf("network not described on %s", c.image)
	}

	left := 10
	for {
		if left <= 0 {
			log.Fatalln("port not available for 10 seconds")
		}
		c, err := net.Dial(nw, net.JoinHostPort(c.host, strconv.Itoa(p)))
		if err == nil {
			c.Close()
			break
		}
		time.Sleep(1 * time.Second)
		left--
	}

	return p
}

// Port returns exposed port in docker host.
func (c *Container) Port(port int) int {
	return c.waitPort(port)
}

// Addr returns exposed address like 127.0.0.1:6379.
func (c *Container) Addr(port int) string {
	exposed := c.waitPort(port)
	return net.JoinHostPort(c.host, strconv.Itoa(exposed))
}

// run command and get result.
func run(name string, args ...string) (out string) {

	cmd := exec.Command(name, args...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("command failed %s %v err:%v", name, args, err)
	}

	if cmd.ProcessState.Success() {
		return strings.TrimSpace(stdout.String())
	}

	log.Fatalf("command execution failed %v", stderr.String())
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
