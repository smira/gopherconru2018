package statsd

import (
	"net"
	"strconv"
	"time"
)

type Client struct {
	sock net.Conn
}

func NewClient(endpoint string) *Client {
	sock, err := net.Dial("udp", endpoint)
	if err != nil {
		panic(err)
	}
	return &Client{sock: sock}
}

// START OMIT
func (c *Client) Incr(stat string, value int64) error {
	buf := append([]byte(stat), ':')
	buf = strconv.AppendInt(buf, value, 10) // HL
	buf = append(buf, []byte("|c")...)
	_, err := c.sock.Write(buf)
	return err

}

// END OMIT

func (c *Client) PrecisionTiming(stat string, delta time.Duration) error {
	buf := append([]byte(stat), ':')
	buf = strconv.AppendFloat(buf, float64(delta)/float64(time.Millisecond), 'f', -1, 64)
	buf = append(buf, []byte("|ms")...)
	_, err := c.sock.Write(buf)
	return err
}

func (c *Client) Close() error {
	return c.sock.Close()
}

func (c *Client) GetLostPackets() int {
	return 0
}
