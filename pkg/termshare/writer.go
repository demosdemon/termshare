package termshare

import (
	"time"
)

type writer struct {
	start time.Time
	t     MessageType
	ch    chan<- *Message
}

func (w *writer) Write(b []byte) (int, error) {
	w.ch <- &Message{
		Offset:      Duration(time.Since(w.start)),
		MessageType: w.t,
		Message:     string(b),
	}
	return len(b), nil
}
