package statsd

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	endpoint string
}

// START OMIT
func NewClient(endpoint string) *Client {
	return &Client{endpoint: endpoint}
}

func (c *Client) Incr(stat string, value int64) error {
	sock, err := net.Dial("udp", c.endpoint) // HL
	if err != nil {
		return err
	}

	_, err = sock.Write([]byte(fmt.Sprintf("%s:%d|c", stat, value))) // HL
	if err != nil {
		return err
	}

	return sock.Close()
}

// END OMIT

func (c *Client) PrecisionTiming(stat string, delta time.Duration) error {
	sock, err := net.Dial("udp", c.endpoint)
	if err != nil {
		return err
	}

	_, err = sock.Write([]byte(fmt.Sprintf("%s:%.3f|ms", stat, float64(delta)/float64(time.Millisecond))))
	if err != nil {
		return err
	}

	return sock.Close()
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) GetLostPackets() int {
	return 0
}
