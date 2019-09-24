// Craig Tomkow
// July 24, 2019

// Modified from copyrighted work (Mozilla Public License 2.0) by Bram Vandenbogaerde (https://github.com/bramvdbogaerde/go-scp)

package remote

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io"
	"path"
	"sync"
	"time"
)

func (sh *SSH) CopyBytes(byteBuffer []byte, filename string, workingDir string, permissions string) error {

	byteReader := bytes.NewReader(byteBuffer)
	return sh.copy(byteReader, workingDir+filename, permissions, int64(len(byteBuffer)))
}

func (sh *SSH) copy(r io.Reader, absolutePath string, permissions string, size int64) error {

	filename := path.Base(absolutePath)
	directory := path.Dir(absolutePath)

	wg := sync.WaitGroup{}
	wg.Add(2)

	errCh := make(chan error, 2)

	go func() {
		defer wg.Done()
		w, err := sh.session.StdinPipe()
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
		err := sh.session.Run(fmt.Sprintf("%s -qt %s", "/usr/bin/scp", directory))
		if err != nil {
			errCh <- err
			return
		}
	}()

	// TODO: remove static timeout
	// time.Duration is in nanoseconds. Default is 1000 seconds
	if waitTimeout(&wg, time.Duration(1000000000000)) {
		return errors.New("timeout when upload files")
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
