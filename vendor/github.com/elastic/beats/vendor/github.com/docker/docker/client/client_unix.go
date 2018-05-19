// Modified to support AIX by daniel.balko@wuerth-it.com

// +build linux freebsd solaris openbsd darwin netbsd aix

package client

// DefaultDockerHost defines os specific default if DOCKER_HOST is unset
const DefaultDockerHost = "unix:///var/run/docker.sock"
