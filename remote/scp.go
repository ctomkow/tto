// Craig Tomkow
// July 24, 2019

// Modified from copyrighted work (Mozilla Public License 2.0) by Bram Vandenbogaerde (https://github.com/bramvdbogaerde/go-scp)

package remote

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

func (sc *SSH) CopyFile(filename string, remotePath string, permissions string) error {

	fd, err := os.Open(filename)
	if err != nil {
		return err
	}

	contentBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}
	byteReader := bytes.NewReader(contentBytes)

	return sc.copy(byteReader, remotePath + filename, permissions, int64(len(contentBytes)))

}

func (sc *SSH) copy(r io.Reader, absolutePath string, permissions string, size int64) error {

	filename := path.Base(absolutePath)
	directory := path.Dir(absolutePath)

	wg := sync.WaitGroup{}
	wg.Add(2)

	errCh := make(chan error, 2)

	go func() {
		defer wg.Done()
		w, err := sc.session.StdinPipe()
		if err != nil {
			errCh <- err
			return
		}
		defer w.Close()

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
		err := sc.session.Run(fmt.Sprintf("%s -qt %s", "/usr/bin/scp", directory))
		if err != nil {
			errCh <- err
			return
		}
	}()

	// time.Duration is in nanoseconds. Default is 100 seconds
	if waitTimeout(&wg, time.Duration(100000000000)) {
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
