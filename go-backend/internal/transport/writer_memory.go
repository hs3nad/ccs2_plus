package transport

import "context"

type MemoryWriter struct {
	Frames [][]byte
}

func (w *MemoryWriter) WriteFrame(_ context.Context, frame []byte) error {
	copied := append([]byte(nil), frame...)
	w.Frames = append(w.Frames, copied)
	return nil
}
