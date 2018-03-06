package statsd

import (
	"net"
	"strconv"
	"sync"
	"time"
)

const PacketSize = 1400

type Client struct {
	sock     net.Conn
	buf      []byte
	wg       sync.WaitGroup
	mutex    sync.Mutex
	shutdown chan struct{}
}

func NewClient(endpoint string) *Client {
	sock, err := net.Dial("udp", endpoint)
	if err != nil {
		panic(err)
	}
	c := &Client{
		sock:     sock,
		buf:      make([]byte, 0, PacketSize),
		shutdown: make(chan struct{}),
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-time.After(100 * time.Millisecond):
				c.mutex.Lock()
				c.checkBuf(PacketSize)
				c.mutex.Unlock()
			case <-c.shutdown:
				return
			}
		}
	}()

	return c
}

func (c *Client) checkBuf(required int) error {
	if len(c.buf)+required > cap(c.buf) && len(c.buf) > 0 {
		_, err := c.sock.Write(c.buf[0 : len(c.buf)-1]) // chop off last \n
		c.buf = c.buf[0:0]
		return err
	}

	return nil
}

// START OMIT
func (c *Client) Incr(stat string, value int64) error {
	c.mutex.Lock()         // HL
	defer c.mutex.Unlock() // HL

	if err := c.checkBuf(len(stat) + 15); err != nil {
		return err
	}

	c.buf = append(c.buf, []byte(stat)...)
	c.buf = append(c.buf, ':')
	c.buf = strconv.AppendInt(c.buf, value, 10)
	c.buf = append(c.buf, []byte("|c\n")...)
	return nil
}

// END OMIT

func (c *Client) PrecisionTiming(stat string, delta time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.checkBuf(len(stat) + 25); err != nil {
		return err
	}

	c.buf = append(c.buf, []byte(stat)...)
	c.buf = append(c.buf, ':')
	c.buf = strconv.AppendFloat(c.buf, float64(delta)/float64(time.Millisecond), 'f', -1, 64)
	c.buf = append(c.buf, []byte("|ms\n")...)
	return nil
}

func (c *Client) Close() error {
	c.mutex.Lock()
	c.checkBuf(PacketSize)
	c.mutex.Unlock()
	close(c.shutdown)
	c.wg.Wait()
	return c.sock.Close()
}

func (c *Client) GetLostPackets() int {
	return 0
}
