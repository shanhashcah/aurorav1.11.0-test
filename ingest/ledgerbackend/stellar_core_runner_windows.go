// +build windows

package ledgerbackend

import (
	"fmt"
	"io"
	"os"

	"github.com/Microsoft/go-winio"
)

// Windows-specific methods for the hcnetCoreRunner type.

func (c *hcnetCoreRunner) getPipeName() string {
	return fmt.Sprintf(`\\.\pipe\%s`, c.nonce)
}

func (c *hcnetCoreRunner) start() (io.Reader, error) {
	// First set up the server pipe.
	listener, err := winio.ListenPipe(c.getPipeName(), nil)
	if err != nil {
		return io.Reader(nil), err
	}

	// Then start the process.
	err = c.cmd.Start()
	if err != nil {
		return io.Reader(nil), err
	}

	c.wg.Add(1)
	go func() {
		select {
		case c.processExit <- c.cmd.Wait():
		case <-c.shutdown:
		}
		c.wg.Done()
	}()

	// Then accept on the server end.
	connection, err := listener.Accept()
	if err != nil {
		return connection, err
	}

	return connection, nil
}

func (c *hcnetCoreRunner) processIsAlive() bool {
	if c.cmd == nil {
		return false
	}
	if c.cmd.Process == nil {
		return false
	}
	p, err := os.FindProcess(c.cmd.Process.Pid)
	if err != nil || p == nil {
		return false
	}
	return true
}
