package transport

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type SerialWriter struct {
	mu         sync.Mutex
	file       *os.File
	writeDelay time.Duration
}

const (
	ioctlTCGETS      = 0x5401
	ioctlTCSETS      = 0x5402
	cflagCRTSCTS     = 0x80000000
)

func NewSerialWriter(device string, baudRate int, writeDelay time.Duration) (*SerialWriter, error) {
	file, err := os.OpenFile(device, os.O_RDWR|syscall.O_NOCTTY|syscall.O_SYNC, 0)
	if err != nil {
		return nil, fmt.Errorf("open serial device: %w", err)
	}

	if err := configureSerial(int(file.Fd()), baudRate); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("configure serial device: %w", err)
	}

	return &SerialWriter{
		file:       file,
		writeDelay: writeDelay,
	}, nil
}

func (w *SerialWriter) WriteFrame(ctx context.Context, frame []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.file.Write(frame); err != nil {
		return fmt.Errorf("write serial frame: %w", err)
	}

	if w.writeDelay > 0 {
		timer := time.NewTimer(w.writeDelay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}

	return nil
}

func (w *SerialWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	return w.file.Close()
}

func configureSerial(fd int, baudRate int) error {
	termios, err := ioctlGetTermios(fd, ioctlTCGETS)
	if err != nil {
		return err
	}

	termios.Iflag = 0
	termios.Oflag = 0
	termios.Lflag = 0
	termios.Cflag &^= syscall.PARENB | syscall.CSTOPB | syscall.CSIZE | cflagCRTSCTS
	termios.Cflag |= syscall.CS8 | syscall.CREAD | syscall.CLOCAL
	termios.Cc[syscall.VMIN] = 1
	termios.Cc[syscall.VTIME] = 0

	speed, err := baudToUnix(baudRate)
	if err != nil {
		return err
	}

	termios.Ispeed = speed
	termios.Ospeed = speed

	return ioctlSetTermios(fd, ioctlTCSETS, termios)
}

func baudToUnix(baudRate int) (uint32, error) {
	switch baudRate {
	case 1200:
		return syscall.B1200, nil
	case 2400:
		return syscall.B2400, nil
	case 4800:
		return syscall.B4800, nil
	case 9600:
		return syscall.B9600, nil
	case 19200:
		return syscall.B19200, nil
	case 38400:
		return syscall.B38400, nil
	default:
		return 0, fmt.Errorf("unsupported baud rate: %d", baudRate)
	}
}

func ioctlGetTermios(fd int, req uintptr) (*syscall.Termios, error) {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), req, uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		return nil, errno
	}
	return &termios, nil
}

func ioctlSetTermios(fd int, req uintptr, termios *syscall.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), req, uintptr(unsafe.Pointer(termios)))
	if errno != 0 {
		return errno
	}
	return nil
}
