package statsd

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	sock net.Conn
}

// START OMIT
func NewClient(endpoint string) *Client {
	sock, err := net.Dial("udp", endpoint)
	if err != nil {
		panic(err)
	}
	return &Client{sock: sock}
}

func (c *Client) Incr(stat string, value int64) error {
	_, err := c.sock.Write([]byte(fmt.Sprintf("%s:%d|c", stat, value)))
	return err

}

// END OMIT

func (c *Client) PrecisionTiming(stat string, delta time.Duration) error {
	_, err := c.sock.Write([]byte(fmt.Sprintf("%s:%.3f|ms", stat, float64(delta)/float64(time.Millisecond))))
	return err
}

func (c *Client) Close() error {
	return c.sock.Close()
}

func (c *Client) GetLostPackets() int {
	return 0
}
