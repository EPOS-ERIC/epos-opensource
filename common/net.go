package common

import (
	"fmt"
	"net"
)

func FindFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	addr := l.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("error dialing udp: %w", err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP.String(), nil
}

func IsPortFree(port int) (bool, error) {
	if port <= 0 || port > 65535 {
		return false, fmt.Errorf("invalid port %d: port must be between 1 and 65535", port)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false, nil
	}

	closeErr := ln.Close()
	if closeErr != nil {
		return true, fmt.Errorf("error closing listener: %w", closeErr)
	}

	return true, nil
}
