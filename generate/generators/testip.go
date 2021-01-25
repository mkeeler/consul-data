package generators

import (
	"math/rand"
	"net"
)

func randomTestingIP() (net.IP, error) {
	octets := make([]byte, 3)

	// its guaranteed to succeed and return len of the array so no
	// need to check the return value
	rand.Read(octets)

	// 198.18.0.0 - 198.19.255.255 is reserved for testing purposes by the IANA
	return net.IPv4(198, 18+(octets[0]&0x1), octets[1], octets[2]), nil
}

func RandomTestingIPGenerator() IPGenerator {
	return randomTestingIP
}
