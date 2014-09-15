package upload

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"
)

type UploadToLocalFile struct {
	lock *sync.RWMutex

	lock_state *sync.Mutex
	state      int

	pool UploaderPool
	id   string

	boundToSocketHandler bool

	dir      string
	path     string
	fd       *os.File
	filePos  int64
	fileSize int64

	signalFinishURL        *url.URL
	signalFinishSecret     string
	removeFileWhenFinished bool
	chHandoverWait         chan error
	chHandoverDone         chan struct{}

	creationTime    time.Time
	lastActionTime  time.Time
	idleTimeout     time.Duration
	canResetTimeout bool
	chResetTimeout  chan time.Duration
	// channel is closed when timeout has been triggered, i.e. nothing should
	// be sent over chResetTimeout any more
	chHandleTimeoutClosed chan struct{}
}

func NewUploadToLocalFile(pool UploaderPool, storageDir string,
	signalFinishURL *url.URL, removeFileWhenFinished bool,
	signalFinishSecret string, idleTimeout time.Duration) Uploader {

	u := new(UploadToLocalFile)
	u.lock = new(sync.RWMutex)
	u.lock_state = new(sync.Mutex)
	u.pool = pool
	u.signalFinishURL = signalFinishURL
	u.signalFinishSecret = signalFinishSecret
	u.removeFileWhenFinished = removeFileWhenFinished
	u.boundToSocketHandler = false
	u.dir = storageDir
	u.chHandoverWait = make(chan error)
	u.chHandoverDone = make(chan struct{})

	u.creationTime = time.Now()
	u.lastActionTime = u.creationTime
	u.idleTimeout = idleTimeout
	u.canResetTimeout = true
	u.chResetTimeout = make(chan time.Duration)
	u.chHandleTimeoutClosed = make(chan struct{})
	go u.goHandleTimeout()

	u.id = pool.Put(u)

	return u
}

func (u *UploadToLocalFile) GetState() int {
	u.lock_state.Lock()
	defer u.lock_state.Unlock()
	return u.state
}

func (u *UploadToLocalFile) GetSignalFinishURL() *url.URL {
	u.lock.RLock()
	defer u.lock.RUnlock()
	ret := *u.signalFinishURL
	return &ret
}

func (u *UploadToLocalFile) GetCreationTime() time.Time {
	u.lock.RLock()
	defer u.lock.RUnlock()
	ret := u.creationTime
	return ret
}

func (u *UploadToLocalFile) GetIdleDuration() time.Duration {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return time.Since(u.lastActionTime)
}

func (u *UploadToLocalFile) ResetTimeout(d time.Duration) Uploader {
	u.lock.Lock()

	// make sure that we are in a state where doing this makes any sense
	u.lock_state.Lock()
	if u.state >= StateCancelled {
		u.lock_state.Unlock()
		u.lock.Unlock()
		return u
	}
	u.lock_state.Unlock()

	u.resetTimeout(d)
	u.lock.Unlock()
	return u
}

// resetTimeout resets the timeout to the given duration and updates
// lastActionTime. u.lock must be held!
func (u *UploadToLocalFile) resetTimeout(d time.Duration) {
	if u.canResetTimeout {
		select {
		case u.chResetTimeout <- d:
		case <-u.chHandleTimeoutClosed:
		}
		u.idleTimeout = d
		u.lastActionTime = time.Now()
	}
}

// goHandleTimeout is a goroutine that waits for the timeout to happen, and
// cancels the upload when the timeout happens.  The goroutine starts with
// u.idleTimeout as timeout.  goHandleTimeout waits for timeout or a call to
// resetTimeut, whichever happens first. A new timeout duration of 0 disables
// the timeout.
// The goroutine terminates when u.chHandleTimeout is closed. CleanUp will do
// that.
func (u *UploadToLocalFile) goHandleTimeout() {
	timer := time.NewTimer(u.idleTimeout)

	for {
		select {
		case d := <-u.chResetTimeout:
			// stop or reset ("prime") timer
			if d == 0 {
				timer.Stop()
			} else {
				timer.Reset(d)
			}
		case <-timer.C:
			// cancel and clean up upload.
			u.lock.Lock()
			u.canResetTimeout = false
			u.lock.Unlock()
			log.Printf("upload %s timed out", u.GetId())
			u.Cancel(true, "upload timed out", 5*time.Second)
			u.CleanUp()
		case <-u.chHandleTimeoutClosed:
			// uploader is done and cleaned up
			return
		}
	}
}

func (u *UploadToLocalFile) GetId() string {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.id
}

func (u *UploadToLocalFile) GetFilePos() int64 {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.filePos
}

func (u *UploadToLocalFile) GetFileSize() int64 {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.fileSize
}

func (u *UploadToLocalFile) SetFileSize(size int64) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.state != StateInit {
		return errors.New("too late to call SetFileSize")
	}

	u.fileSize = size
	u.resetTimeout(u.idleTimeout)
	return nil
}

func (u *UploadToLocalFile) BindToSocketHandler() error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.boundToSocketHandler {
		return errors.New("Bound to some socket handler already!")
	}
	u.boundToSocketHandler = true
	u.resetTimeout(u.idleTimeout)
	return nil
}

func (u *UploadToLocalFile) UnbindFromSocketHandler() error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if !u.boundToSocketHandler {
		return errors.New("not bound to any socket handler")
	}
	u.boundToSocketHandler = false
	u.resetTimeout(u.idleTimeout)
	return nil
}

func (u *UploadToLocalFile) ConsumeFileChunk(chunk []byte) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	defer u.resetTimeout(u.idleTimeout)

	// quite a bit of "state business" follows.
	u.lock_state.Lock()

	// make new file if we have to
	if u.state == StateInit {
		u.path = path.Join(u.dir, fmt.Sprintf("%s.part", u.id))
		//log.Printf("creating file %s", u.path)
		fd, err := os.Create(u.path)
		if err != nil {
			u.lock_state.Unlock()
			return errors.New("Could not create file! file system full?")
		}
		u.fd = fd
	}

	// if we are resuming an upload, open the file we already have
	if u.state == StatePaused {
		fd, err := os.OpenFile(u.path, os.O_RDWR, 0666)
		if err != nil {
			u.lock_state.Unlock()
			return errors.New("Could not re-open file! Is it gone?")
		}
		u.fd = fd
		_, _ = u.fd.Seek(0, os.SEEK_END)
	}

	// make sure we are in a legal state to proceed (i.e., not in any of the "we're
	// done uploading" states)
	if u.state > StatePaused {
		u.lock_state.Unlock()
		return errors.New("upload is in no state for this. might be cancelled.")
	}

	// set state to "uploading"
	if u.state != StateUploading {
		u.state = StateUploading
	}

	// "state business" ends.
	u.lock_state.Unlock()

	// assert that fileSize will not be exceeded
	if u.filePos+int64(len(chunk)) > u.fileSize {
		return errors.New("File would get larger than declared")
	}

	// write! and if there was a problem, undo the write
	bytesWritten, err := u.fd.Write(chunk)
	if err != nil {
		u.fd.Truncate(u.filePos)
		_, _ = u.fd.Seek(0, os.SEEK_END)
		return err
	}
	u.filePos += int64(bytesWritten)

	// if file is complete, close and rename it
	if u.filePos == u.fileSize {
		u.fd.Close()
		u.fd = nil
		newName := u.path[:len(u.path)-5]
		err = os.Rename(u.path, newName)
		if err != nil {
			return err
		}
		u.path = newName
	}

	return nil
}

func (u *UploadToLocalFile) Pause() (err error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	// assert that we are in a legal state, set state to paused
	u.lock_state.Lock()
	if u.state != StateUploading && u.state != StatePaused {
		err = errors.New("can't pause now")
	} else {
		u.state = StatePaused
	}
	u.lock_state.Unlock()
	if err != nil {
		return
	}

	// close the file
	if u.fd != nil {
		u.fd.Close()
		u.fd = nil
	}

	u.resetTimeout(u.idleTimeout)
	return nil
}

func (u *UploadToLocalFile) HandFileToApp(reqTimeout time.Duration,
	respTimeout time.Duration) (ch_ret chan error) {
	u.lock.RLock()
	ch_ret = u.chHandoverWait
	u.lock.RUnlock()

	// figure out whether we have to do anything (we might have been called
	// before or we might be in a wrong state)
	u.lock_state.Lock()
	run := (u.state < StateHandingOver)
	if run {
		u.state = StateHandingOver
	}
	u.lock_state.Unlock()

	if !run {
		return
	}

	go func() {
		htclient := new(http.Client)
		htclient.Timeout = reqTimeout

		// signal app backend that we are done
		v := url.Values{}
		v.Set("id", u.id)
		v.Set("filename", u.path)
		v.Set("secret", u.signalFinishSecret)
		v.Set("cancelled", "no")
		v.Set("cancelReason", "")
		u.lock.Lock()
		u.resetTimeout(u.idleTimeout)
		u.lock.Unlock()
		resp, err := htclient.PostForm(u.signalFinishURL.String(), v) // this takes time
		u.lock.Lock()
		u.resetTimeout(u.idleTimeout)
		u.lock.Unlock()

		// set error if http went through but we got a bad http status back
		if err == nil && resp.StatusCode != 200 {
			//log.Printf("Got bad http status on handover: %s", resp.Status)
			err = fmt.Errorf("Got bad http status on handover: %s", resp.Status)
		}

		// read (first 4 bytes of) response body if we can
		respBody := []byte(nil)
		if err == nil {
			if resp.ContentLength > -1 {
				readLimit := resp.ContentLength
				if readLimit > 4 {
					readLimit = 4
				}
				respBody = make([]byte, readLimit)
				resp.Body.Read(respBody)
				resp.Body.Close()
			}
		}
		respStr := string(respBody)
		//log.Printf("Got response from app backend: %s", respStr)

		// response is "done"? yay, we'll be done. response is "wait"? we'll wait...
		wait := false
		if err == nil {
			if respStr == "wait" {
				wait = true
			} else if respStr == "done" {
				wait = false
			} else {
				err = errors.New("don't understand reply from app backend")
			}
		}

		// bail on error
		if err != nil {
			// set state to cancelled, propagate error if anybody listens, return
			u.lock_state.Lock()
			u.state = StateCancelled
			u.lock_state.Unlock()
			select {
			case ch_ret <- err:
			case <-time.After(1 * time.Second):
			}
			close(ch_ret)
			return
		}

		// wait if we have to
		if wait {
			log.Printf("wait for app backend")
			select {
			case <-u.chHandoverDone:
				u.lock.Lock()
				u.resetTimeout(u.idleTimeout)
				u.lock.Unlock()
			case <-time.After(respTimeout):
				err = errors.New("Timed out waiting for app backend to retrieve the file")
			}
			log.Printf("wait done")
		}

		// update state
		u.lock_state.Lock()
		if err == nil {
			u.state = StateFinished
		} else {
			u.state = StateCancelled
		}
		u.lock_state.Unlock()

		// try to send error (likely nil) over return channel, then close it
		select {
		case ch_ret <- err:
		case <-time.After(1 * time.Second):
		}
		close(ch_ret)
	}()
	return
}

// called by web app backend to signal that it is done retrieving
// the uploaded file.
func (u *UploadToLocalFile) HandoverDone() error {
	u.lock_state.Lock()
	if u.state != StateHandingOver {
		u.lock_state.Unlock()
		return errors.New("uploader is not in 'handing over' state")
	}
	u.lock_state.Unlock()

	select {
	case u.chHandoverDone <- struct{}{}:
		return nil
	case <-time.After(1 * time.Second):
		return errors.New("no waiting handover routine")
	}
}

func (u *UploadToLocalFile) Cancel(tellAppBackend bool, reason string,
	reqTimeout time.Duration) error {
	u.lock.Lock()

	// set state to cancel if we can
	u.lock_state.Lock()
	alreadyCancelled := (u.state == StateCancelled)
	canCancel := (u.state < StateHandingOver)
	if canCancel {
		u.state = StateCancelled
	}
	u.lock_state.Unlock()

	// return nil if we are already cancelled, an error if we can't cancel
	if alreadyCancelled {
		u.lock.Unlock()
		return nil
	} else if !canCancel {
		u.lock.Unlock()
		return errors.New("too late to cancel")
	}

	// close file if it is open
	if u.fd != nil {
		u.fd.Close()
		u.fd = nil
	}
	// delete file if we already have one
	if u.path != "" {
		os.Remove(u.path)
	}

	u.resetTimeout(u.idleTimeout)

	// tell app backend that we have cancelled if we have to. We don't need to
	// hold the lock for this.
	signalFinishSecret := u.signalFinishSecret
	signalFinishURL := u.signalFinishURL
	id := u.id
	u.lock.Unlock()

	htclient := new(http.Client)
	htclient.Timeout = reqTimeout

	v := url.Values{}
	v.Set("id", id)
	v.Set("filename", "")
	v.Set("secret", signalFinishSecret)
	v.Set("cancelled", "yes")
	v.Set("cancelReason", reason)
	resp, err := htclient.PostForm(signalFinishURL.String(), v) // this takes time

	// set error if http request didn't work
	if err != nil {
		err = fmt.Errorf("http request to app backend at %s failed",
			signalFinishURL.String())
	} else {
		// set error if http went through but we got a bad http status back
		if resp.StatusCode != 200 {
			err = fmt.Errorf("Got bad http status on handover: %s", resp.Status)
		}

		// we don't care what's in the body of the response
		if resp.Body != nil {
			resp.Body.Close()
		}
	}

	u.lock.Lock()
	u.resetTimeout(u.idleTimeout)
	u.lock.Unlock()
	return err
}

func (u *UploadToLocalFile) CleanUp() (err error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	// make sure that we are in a valid state (cancelled or finished)
	u.lock_state.Lock()
	if u.state <= StateHandingOver {
		u.lock_state.Unlock()
		err = fmt.Errorf("It's too early to call CleanUp")
		return
	}
	if u.state == StateCleanedUp {
		u.lock_state.Unlock()
		return
	}
	u.lock_state.Unlock()

	// delete file if we have to
	if u.removeFileWhenFinished && u.state != StateCancelled {
		err = os.Remove(u.path)
		if err != nil {
			log.Printf("could not remove %s during cleanup", u.path)
		}
	}

	// remove ourselves from uploader pool
	u.pool.Remove(u.id)

	// set state to 'cleaned up'
	u.lock_state.Lock()
	u.state = StateCleanedUp
	u.lock_state.Unlock()

	// make sure the timeout handling goroutine terminates, and that calls to
	// ResetTimeout return
	close(u.chHandleTimeoutClosed)

	return
}
