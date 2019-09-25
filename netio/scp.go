// Craig Tomkow
// July 24, 2019

// Modified from copyrighted work (Mozilla Public License 2.0) by Bram Vandenbogaerde (https://github.com/bramvdbogaerde/go-scp)

package netio

import (
	"errors"
	"fmt"
	"github.com/ctomkow/tto/exec"
	"github.com/ctomkow/tto/net"
	"github.com/golang/glog"
	"io"
	"path"
	"sync"
	"time"
)

func Copy(byteBuffer *io.ReadCloser, filename string, workingDir string, permissions string, ex *exec.Exec, sh *net.SSH) error {

	// ensure a new session is created before acting!
	if err := sh.NewSession(); err != nil {
		return err
	}

	// since i don't know the size of the dump, set a static max to 100GB (107374182400 bytes)
	if err := copy(*byteBuffer, workingDir+filename, permissions, 107374182400, ex, sh); err != nil {
		glog.Error(err)
	}

	return nil
}

func copy(r io.ReadCloser, absolutePath string, permissions string, size int64, ex *exec.Exec, sh *net.SSH) error {

	filename := path.Base(absolutePath)
	directory := path.Dir(absolutePath)

	wg := sync.WaitGroup{}
	wg.Add(2)

	errCh := make(chan error, 2)

	go func() {
		defer wg.Done()
		w, err := sh.Session.StdinPipe()
		if err != nil {
			errCh <- err
			return
		}
		defer func() {
			if err := w.Close(); err != nil {
				glog.Exit(err)
			}
		}()

		_, err = fmt.Fprintln(w, "C"+permissions, size, filename)
		if err != nil {
			errCh <- err
			return
		}

		_, err = io.Copy(w, r)
		if err != nil {
			errCh <- err
			return
		}

		_, err = fmt.Fprint(w, "\x00")
		if err != nil {
			errCh <- err
			return
		}

	}()

	go func() {
		defer wg.Done()
		err := sh.Session.Run(fmt.Sprintf("%s -qt %s", "/usr/bin/scp", directory))
		if err != nil {
			// The SCP process is existing with code 1 because the session is being forcefully closed after transfer is complete
			//   SCP would properly close if we specify a correct file size, but we don't know that because we are streaming mysqldump
			//   Therefore it is set to a max of 100GB
			//   Consequently, we cannot handle an error case here :/

		}
	}()

	// time.Duration is in nanoseconds. Default is 1000 seconds
	if waitTimeout(&wg, time.Duration(1000000000000)) {
		return errors.New("timeout when upload files")
	}

	if err:= ex.Cmd.Wait(); err != nil {
		return err
	}

	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
