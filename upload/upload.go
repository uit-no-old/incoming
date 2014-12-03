/*
Incoming!! uploader module and interface

Copyright (C) 2014 Lars Tiede, University of Tromsø - The Arctic University of Norway


This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package upload

import (
	"net/url"
	"time"
)

// InitModule initializes the upload module. At present, this is only clearing
// the local file uploader's storage directory, which we have to do in case the
// app was shut down while uploads were running.
func InitModule(storageDir string) error {
	return initStorageDir(storageDir)
}

const (
	StateInit = iota
	StateUploading
	StatePaused
	StateHandingOver
	StateCancelled
	StateFinished
	StateCleanedUp
)

type Uploader interface {
	// We allow only one active socket handler per upload. BindToSocketHandler
	// allocates an uploader to a socket handler.
	BindToSocketHandler() error

	// UnbindFromSocketHandler 'deallocates' an uploader from a socket handler.
	UnbindFromSocketHandler() error

	// SetFileSize should be called once before any chunks are uploaded.
	SetFileSize(int64) error

	// SetFileName should be called once before any chunks are uploaded. The
	// name is the name of the file as reported by the browser. It is not the
	// name Incoming!! should use internally.
	SetFileName(string) error

	// GetFileSize returns the size of the file that is being uploaded.
	GetFileSize() int64

	// GetFileName returns the name of the file, as reported by the browser.
	// This is not necessarily the actual file name Incoming!! uses internally.
	GetFileName() string

	// GetFilePos returns the current cursor position within the uploaded
	// file, i.e. how many bytes have been uploaded already.
	GetFilePos() int64

	// GetSignalFinishURL returns the web app backend URL the uploader POSTs to
	// when the upload is finished.
	GetSignalFinishURL() *url.URL

	// GetBackendSecret returns the secret string that the web app backend has
	// given to Incoming!! when requesting an upload ticket. The returned
	// string might be empty, in which case there is no secret string.
	GetBackendSecret() string

	// GetId returns the (textual) ID of the upload.
	GetId() string

	// GetState queries the upload state. It never takes long to return.
	GetState() int

	// ConsumeFileChunk synchronously stores the next file chunk to whichever
	// store the implementation uses.
	// An error is returned if the operation fails. In that case, the write
	// operation 'never happened'. The upload does not cancel automatically.
	ConsumeFileChunk([]byte) error

	// HandFileToApp asynchronously notifies the app backend that a file with a
	// certain id has arrived, and that the app backend can fetch / move / copy
	// it. It then optionally waits until the app backend is finished obtaining
	// the file (whether this wait happens is decided by the app backend).
	// After all of that is done, HandFileToApp sends an error object to the
	// channel the function returns. It's fine if nobody is listening.  When
	// everything is done successfully, the state of the upload is 'finished'.
	//
	// If handover is not successful, the returned error object is not nil, and
	// the upload's state is "cancelled". CleanUp() is not automatically
	// called.
	//
	// HandFileToApp can be called several times while or even after its
	// internal goroutine is running. It will always return the same channel,
	// and write and close that channel eventually. That is to say, for each
	// Uploader, HandFileToApp's functionality runs exactly once.
	//
	// the two parameters are timeouts for the request to the app backend,
	// and waiting for the confirmation request (if there will be any)
	HandFileToApp(time.Duration, time.Duration) chan error

	// HandoverDone should be called by the app backend when it is finished
	// obtaining the file.
	// error is not nil if there was no HandFileToApp routine running, or
	// if the upload was not in the "hand over file" state.
	HandoverDone() error

	// Cancel ends the upload. No new chunks will be accepted.  The first
	// parameter determines whether the app backend should be notified or not.
	// This should be set to true unless Cancel() is called from the app
	// backend itself. The second parameter can hold an error message that will
	// be sent to the app backend. The third parameter specifies a timeout for
	// the http request to the app backend (0 for no timeout). Call Cancel() in
	// its own goroutine if you don't want to wait for it to finish telling the
	// web backend.
	//
	// Cancel() does not automatically call CleanUp().
	//
	// The returned error is not nil either if telling the web app backend was
	// not successful or if the upload can't be cancelled because it is already
	// too far in the process (i.e., if it is in one of the following states:
	// handing over, finished, cleaned up). Cancelling a cancelled upload is a
	// no-op.
	Cancel(bool, string, time.Duration) error

	// Pause can be called to pause an upload while the upload is in progress.
	// Pause will make an Uploader close the file or whatever else it writes
	// to, and open it again when ConsumeFileChunk is called again (which
	// "unpauses" the upload). The idle timeout is not affected by Pause, so in
	// practice it makes (for now at least) little difference whether you call
	// Pause, or whether you not call Pause and just let some time pass before
	// calling ConsumeFileChunk again. This behavior is likely to change in the
	// future.
	Pause() error

	// CleanUp cleans up after a finished or cancelled upload. It removes
	// temporary data, and removes the uploader from the uploader pool.
	CleanUp() error

	// GetCreationTime returns the time when the uploader was created.
	GetCreationTime() time.Time

	// GetIdleDuration returns the duration since the upload has been idle.
	// The 'last action' timestamp is updated on most function calls, and also
	// during a few of the asynchronous goroutines, for example when handingthe
	// file over to the web app backend.
	GetIdleDuration() time.Duration

	// ResetTimeout sets or resets the timeout. This happens automatically in
	// several functions, so you need to call this one only if you want to
	// explicitly set or reset the timeout without doing anything else. A
	// duration of 0 disables the timeout.
	ResetTimeout(time.Duration) Uploader
}
