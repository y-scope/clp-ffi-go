package ir

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"

	"github.com/pierrec/lz4/v4"
)

// filterStore manages an LZ4-compressed temporary file for persisting observed
// values. Values are stored in a length-prefixed binary format
// ([4-byte LE length][value bytes]) to support values containing arbitrary
// bytes including newlines. The file uses concatenated LZ4 frames so the writer
// can be closed and reopened across rebuilds. The temp file is created lazily on
// first write and automatically deleted when [close] is called.
type filterStore struct {
	tmpFile *os.File
	lz4w    *lz4.Writer
	bufw    *bufio.Writer
}

// ensureOpen creates the temp file and LZ4 writer on first use.
func (fs *filterStore) ensureOpen() error {
	if nil == fs.tmpFile {
		var err error
		fs.tmpFile, err = os.CreateTemp("", "clp-ffi-go-filter-*.tmp")
		if nil != err {
			return err
		}
	}
	if nil == fs.lz4w {
		fs.lz4w = lz4.NewWriter(fs.tmpFile)
		fs.bufw = bufio.NewWriter(fs.lz4w)
	}
	return nil
}

// append writes a length-prefixed value to the LZ4-compressed temp file.
func (fs *filterStore) append(data []byte) error {
	if err := fs.ensureOpen(); nil != err {
		return err
	}
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(data)))
	if _, err := fs.bufw.Write(lenBuf[:]); nil != err {
		return err
	}
	if _, err := fs.bufw.Write(data); nil != err {
		return err
	}
	return nil
}

// closeLz4Writer flushes and closes the current LZ4 frame, finalizing it for
// reading. The underlying file remains open.
func (fs *filterStore) closeLz4Writer() error {
	if nil == fs.lz4w {
		return nil
	}
	if err := fs.bufw.Flush(); nil != err {
		return err
	}
	if err := fs.lz4w.Close(); nil != err {
		return err
	}
	fs.lz4w = nil
	fs.bufw = nil
	return nil
}

// forEach reads all length-prefixed values from the LZ4-compressed temp file
// and calls fn for each value. The LZ4 reader transparently reads across
// concatenated frames. The LZ4 writer must be closed before calling this
// method.
func (fs *filterStore) forEach(fn func([]byte)) error {
	if nil == fs.tmpFile {
		return nil
	}

	if _, err := fs.tmpFile.Seek(0, 0); nil != err {
		return err
	}

	lz4r := lz4.NewReader(fs.tmpFile)
	bufr := bufio.NewReader(lz4r)

	for {
		var lenBuf [4]byte
		if _, err := io.ReadFull(bufr, lenBuf[:]); nil != err {
			if io.EOF == err || io.ErrUnexpectedEOF == err {
				break // end of data or truncated length prefix
			}
			return err
		}
		n := binary.LittleEndian.Uint32(lenBuf[:])
		val := make([]byte, n)
		if _, err := io.ReadFull(bufr, val); nil != err {
			if io.ErrUnexpectedEOF == err {
				break // truncated record — treat as end of valid data
			}
			return err
		}
		fn(val)
	}
	return nil
}

// reopenWriter opens a new LZ4 writer at the current file position for
// continued appending. Must be called after [closeLz4Writer] and any read
// operations.
func (fs *filterStore) reopenWriter() error {
	if nil == fs.tmpFile {
		return nil
	}
	// Seek to end so new writes append after existing frames.
	if _, err := fs.tmpFile.Seek(0, 2); nil != err {
		return err
	}
	fs.lz4w = lz4.NewWriter(fs.tmpFile)
	fs.bufw = bufio.NewWriter(fs.lz4w)
	return nil
}

// close flushes, closes, and removes the temp file. Returns the first error
// encountered but always attempts all cleanup steps.
func (fs *filterStore) close() error {
	var firstErr error
	if nil != fs.lz4w {
		if err := fs.bufw.Flush(); nil != err && nil == firstErr {
			firstErr = err
		}
		if err := fs.lz4w.Close(); nil != err && nil == firstErr {
			firstErr = err
		}
		fs.lz4w = nil
		fs.bufw = nil
	}
	if nil == fs.tmpFile {
		return firstErr
	}
	name := fs.tmpFile.Name()
	if err := fs.tmpFile.Close(); nil != err && nil == firstErr {
		firstErr = err
	}
	fs.tmpFile = nil
	if err := os.Remove(name); nil != err && nil == firstErr {
		firstErr = err
	}
	return firstErr
}
