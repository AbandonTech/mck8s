package tcp

import (
	"io"
	"net"
	"sync"
)

// Proxy two connections forwarding each message to the other connection
func Proxy(c1 net.Conn, c2 net.Conn) error {

	var err error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		_, err2 := io.Copy(c1, c2)
		if err2 != nil {
			err = err2
			return
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		_, err2 := io.Copy(c2, c1)
		if err2 != nil {
			err = err2
			return
		}
		wg.Done()
	}()

	wg.Wait()
	return err
}
