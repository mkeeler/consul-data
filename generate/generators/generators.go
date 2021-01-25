package generators

import "net"

type StringGenerator func() (string, error)

type IPGenerator func() (net.IP, error)
