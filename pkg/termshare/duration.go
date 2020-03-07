package termshare

import (
	"encoding/json"
	"time"
)

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	v := float64(d) / float64(time.Second)
	return json.Marshal(v)
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*d = Duration(v * float64(time.Second))
	return nil
}

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	v := Duration(t.UnixNano())
	return v.MarshalJSON()
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var v Duration
	if err := v.UnmarshalJSON(data); err != nil {
		return err
	}
	*t = Time{Time: time.Unix(0, 0).Add(time.Duration(v))}
	return nil
}
