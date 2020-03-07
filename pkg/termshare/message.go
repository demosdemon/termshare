package termshare

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/demosdemon/termshare/pkg/utils"
)

type MessageType string

const (
	StdinMessage  MessageType = "i"
	StdoutMessage MessageType = "o"
)

type Message struct {
	Offset      Duration
	MessageType MessageType
	Message     string
}

func (m Message) MarshalJSON() ([]byte, error) {
	d := float64(m.Offset) / float64(time.Second)
	x := []interface{}{
		d,
		m.MessageType,
		m.Message,
	}
	return json.Marshal(x)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	var (
		d   Duration
		t   MessageType
		msg string
	)

	if err := utils.ErrorSequence(
		func() error { return expect(dec, '[') },
		func() error { return dec.Decode(&d) },
		func() error { return dec.Decode(&t) },
		func() error { return dec.Decode(&msg) },
		func() error { return expect(dec, ']') },
	); err != nil {
		return err
	}

	*m = Message{
		Offset:      d,
		MessageType: t,
		Message:     msg,
	}
	return nil
}

type Processor func(message Message) (Message, error)

type Processors []Processor

func (p Processors) Apply(msg Message) (Message, error) {
	for _, fn := range p {
		var err error
		msg, err = fn(msg)
		if err != nil {
			return msg, err
		}
	}
	return msg, nil
}

func ToRelativeTime() Processor {
	prev := Duration(0)
	return func(msg Message) (Message, error) {
		d := msg.Offset - prev
		prev = msg.Offset
		msg.Offset = d
		return msg, nil
	}
}

func ToAbsoluteTime() Processor {
	prev := Duration(0)
	return func(msg Message) (Message, error) {
		prev += msg.Offset
		msg.Offset = prev
		return msg, nil
	}
}

func CapRelativeTime(limit Duration) Processor {
	return func(msg Message) (Message, error) {
		if limit > 0 && msg.Offset > limit {
			msg.Offset = limit
		}
		return msg, nil
	}
}

func AdjustSpeed(speed float64) Processor {
	return func(msg Message) (Message, error) {
		if speed > 0 {
			msg.Offset = Duration(float64(msg.Offset) / speed)
		}
		return msg, nil
	}
}

func expect(dec *json.Decoder, d json.Delim) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	if t == d {
		return nil
	}

	return errors.Errorf("unexpected token: %v", t)
}
