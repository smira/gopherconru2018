package statsd

import (
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func setupListener(t *testing.T) (*net.UDPConn, chan []byte) {
	inSocket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP: net.IPv4(127, 0, 0, 1),
	})
	if err != nil {
		t.Error(err)
	}

	received := make(chan []byte)

	go func() {
		for {
			buf := make([]byte, 1500)

			n, err := inSocket.Read(buf)
			if err != nil {
				return
			}

			received <- buf[0:n]
		}

	}()

	return inSocket, received
}

func TestCommands(t *testing.T) {
	inSocket, received := setupListener(t)

	client := NewClient(inSocket.LocalAddr().String())

	compareOutput := func(actions func() error, expected []string) func(*testing.T) {
		return func(t *testing.T) {
			err := actions()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for _, exp := range expected {
				buf := <-received

				if string(buf) != exp {
					t.Errorf("unexpected part received: %#v != %#v", string(buf), exp)
				}
			}
		}
	}

	t.Run("Incr", compareOutput(
		func() error { return client.Incr("req.count", 30) },
		[]string{"req.count:30|c"}))

	t.Run("PrecisionTiming", compareOutput(
		func() error { return client.PrecisionTiming("req.duration", 157356*time.Microsecond) },
		[]string{"req.duration:157.356|ms"}))

	_ = client.Close()
	_ = inSocket.Close()
	close(received)
}

func TestConcurrent(t *testing.T) {
	inSocket, received := setupListener(t)

	client := NewClient(inSocket.LocalAddr().String())

	var totalSent, totalReceived int64

	var wg1, wg2 sync.WaitGroup

	wg1.Add(1)

	go func() {
		for buf := range received {
			for _, part := range strings.Split(string(buf), "\n") {
				i1 := strings.Index(part, ":")
				i2 := strings.Index(part, "|")

				if i1 == -1 || i2 == -1 {
					t.Logf("unaparseable part: %#v", part)
					continue
				}

				count, err := strconv.ParseInt(part[i1+1:i2], 10, 64)
				if err != nil {
					t.Log(err)
					continue
				}

				atomic.AddInt64(&totalReceived, count)
			}
		}

		wg1.Done()
	}()

	workers := 8
	count := 4096

	for i := 0; i < workers; i++ {
		wg2.Add(1)

		go func(i int) {
			for j := 0; j < count; j++ {
				// to simulate real load, sleep a bit in between the stats calls
				time.Sleep(time.Duration(rand.ExpFloat64() * float64(time.Microsecond)))

				increment := i + j
				client.Incr("some.counter", int64(increment))

				atomic.AddInt64(&totalSent, int64(increment))
			}

			wg2.Done()
		}(i)
	}

	wg2.Wait()

	if client.GetLostPackets() > 0 {
		t.Errorf("some packets were lost during the test, results are not valid: %d", client.GetLostPackets())
	}

	_ = client.Close()

	time.Sleep(time.Second)

	_ = inSocket.Close()
	close(received)

	wg1.Wait()

	if atomic.LoadInt64(&totalSent) != atomic.LoadInt64(&totalReceived) {
		t.Errorf("sent != received: %v != %v", totalSent, totalReceived)
	}
}

func BenchmarkClient(b *testing.B) {
	inSocket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP: net.IPv4(127, 0, 0, 1),
	})
	if err != nil {
		b.Error(err)
	}

	go func() {
		buf := make([]byte, 1500)
		for {
			_, err := inSocket.Read(buf)
			if err != nil {
				return
			}
		}

	}()

	client := NewClient(inSocket.LocalAddr().String())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.Incr("number.requests", 33)
		client.PrecisionTiming("response.time.for.some.api", 150*time.Millisecond)
		client.PrecisionTiming("response.time.for.some.api.case1", 150*time.Millisecond)
	}

	_ = client.Close()
	_ = inSocket.Close()
}
