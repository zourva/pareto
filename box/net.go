package box

import (
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
)

// ValidateEndpoint checks the given ep if it's in format of "host:port".
func ValidateEndpoint(ep string) bool {
	if len(ep) == 0 {
		log.Errorln("invalid endpoint address")
		return false
	}

	if strings.Count(ep, ":") > 1 {
		log.Errorln("invalid endpoint address %s, should be the format of host:port")
		return false
	}

	if strings.Index(ep, ":") == 0 {
		log.Errorln("invalid endpoint address %s, should be the format of host:port")
		return false
	}

	return true
}

// ParseUdpAddr parses the given ep, which should be in format "host:port",
// to a three elements tuple <host, port, net.UDPAddr>.
func ParseUdpAddr(endpoint string) (string, int, *net.UDPAddr) {
	ss := strings.Split(endpoint, ":")
	host := ss[0]
	port, _ := strconv.ParseInt(ss[1], 10, 32)

	return host, int(port), &net.UDPAddr{
		IP:   net.ParseIP(ss[0]),
		Port: int(port),
	}
}
