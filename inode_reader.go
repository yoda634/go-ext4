package ext4

import (
	"io"
	"math"

	log "github.com/dsoprea/go-logging"
)

// InodeReader fulfills the `io.Reader` interface to read arbitrary amounts of
// data.
type InodeReader struct {
	en           *ExtentNavigator
	currentBlock []byte
	bytesRead    uint64
	bytesTotal   uint64
}

func NewInodeReader(en *ExtentNavigator) *InodeReader {
	return &InodeReader{
		en:           en,
		currentBlock: make([]byte, 0),
		bytesTotal:   en.inode.Size(),
	}
}

func (ir *InodeReader) Offset() uint64 {
	return ir.bytesRead
}

func (ir *InodeReader) fill() (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if len(ir.currentBlock) == 0 {
		if ir.bytesRead >= ir.bytesTotal {
			return io.EOF
		}

		data, _, err := ir.en.Read(ir.bytesRead)
		log.PanicIf(err)

		ir.currentBlock = data
		ir.bytesRead += uint64(len(data))
	}

	return nil
}

// Read fills the given slice with data and returns an `io.EOF` error with (0)
// bytes when done. (`n`) may be less then `len(p)`.
func (ir *InodeReader) Read(p []byte) (n int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = ir.fill()
	if err == io.EOF {
		return 0, io.EOF
	} else if err != nil {
		log.PanicIf(err)
	}

	// Determine how much of the buffer we can fill.
	currentBytesReadCount := uint64(math.Min(float64(len(ir.currentBlock)), float64(len(p))))

	copy(p, ir.currentBlock[:currentBytesReadCount])
	ir.currentBlock = ir.currentBlock[currentBytesReadCount:]

	return int(currentBytesReadCount), nil
}

// Skip simulates a read but just discards the data.
func (ir *InodeReader) Skip(n uint64) (skipped uint64, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = ir.fill()
	if err == io.EOF {
		return 0, io.EOF
	} else if err != nil {
		log.Panic(err)
	}

	currentBytesReadCount := uint64(math.Min(float64(len(ir.currentBlock)), float64(n)))
	ir.currentBlock = ir.currentBlock[currentBytesReadCount:]

	return currentBytesReadCount, nil
}
