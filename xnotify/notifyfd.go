package xnotify

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

type fNotify struct {
	mux sync.RWMutex
}

// Internal eventMetadata struct, used for fanotify comm
type eventMetadata struct {
	Len         uint32
	Version     uint8
	Reserved    uint8
	MetadataLen uint16
	Mask        uint64
	Fd          int32
	Pid         int32
}

// Event struct returned from NotifyFD.GetEvent
//
// The File member needs to be Closed after usage, to prevent an Fd leak
type EventMetadata struct {
	Len         uint32
	Version     uint8
	Reserved    uint8
	MetadataLen uint16
	Mask        uint64
	File        *os.File
	Pid         int32
}

// Internal response struct, used for fanotify comm
type response struct {
	Fd       int32
	Response uint32
}

// A notify handle, used by all notify functions
type NotifyFD struct {
	fd int32
	f  *os.File
	r  *bufio.Reader
}

func initialize(faflag, openflags int) (*NotifyFD, error) {
	fd, _, errno := syscall.Syscall(syscall.SYS_FANOTIFY_INIT, uintptr(faflag), uintptr(openflags), uintptr(0))
	var err error
	if errno != 0 {
		err = errno
	}

	f := os.NewFile(fd, "")

	return &NotifyFD{fd: int32(fd), f: f, r: bufio.NewReader(f)}, err
}

func (nd *NotifyFD) GetFd() int32 {
	return nd.fd
}

// Get an event from the fanotify handle
func (nd *NotifyFD) GetEvent() (*EventMetadata, error) {
	ev := &eventMetadata{}

	err := binary.Read(nd.r, binary.LittleEndian, ev)
	if err != nil {
		return nil, err
	}

	res := &EventMetadata{ev.Len, ev.Version, ev.Reserved, ev.MetadataLen, ev.Mask, os.NewFile(uintptr(ev.Fd), ""), ev.Pid}

	return res, nil
}

// Send an allow message back to fanotify, used for permission checks
// If allow is set to true, access is granted
func (nd *NotifyFD) Response(ev *EventMetadata, allow bool) error {
	resp := &response{Fd: int32(ev.File.Fd())}

	if allow {
		resp.Response = FAN_ALLOW
	} else {
		resp.Response = FAN_DENY
	}

	return binary.Write(nd.f, binary.LittleEndian, resp)
}

func (nd *NotifyFD) Close() {
	nd.f.Close()
}

// Get an event from the fanotify handle
func (nd *NotifyFD) GetEvents() ([]*EventMetadata, error) {
	var events []*EventMetadata
	buffer := make([]byte, 4800) // should be enough: 24 bytes/event*200 events = 4800 bytes
	n, err := io.ReadAtLeast(nd.r, buffer, FA_EVENT_LEN)
	if err != nil {
		fmt.Printf("[NotifyFD GetEvents] error:%v\n", err)
		return nil, err
	}

	if (n % FA_EVENT_LEN) != 0 {
		fmt.Printf("[NotifyFD GetEvents] incomplete,error:%v\n", err)
	}

	reader := bytes.NewReader(buffer)
	for i := 0; i < (n / FA_EVENT_LEN); i++ {
		ev := &eventMetadata{}
		if err = binary.Read(reader, binary.LittleEndian, ev); err == nil { // only have the EOF error
			if ev.Version == FANOTIFY_METADATA_VERSION {
				if f := os.NewFile(uintptr(ev.Fd), ""); f != nil {
					events = append(events, &EventMetadata{ev.Len, ev.Version, ev.Reserved, ev.MetadataLen, ev.Mask, f, ev.Pid})
				}
			}
		}
	}
	return events, err
}

// Add/Delete/Modify an Fanotify mark
func (nd *NotifyFD) Mark(flags int, mask uint64, dfd int, path string) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_FANOTIFY_MARK, uintptr(nd.f.Fd()), uintptr(flags), uintptr(mask), uintptr(dfd), uintptr(unsafe.Pointer(syscall.StringBytePtr(path))), 0)

	var err error
	if errno != 0 {
		err = errno
	}

	return err
}
