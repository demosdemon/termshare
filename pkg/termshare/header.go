package termshare

type Header struct {
	Version       int               `json:"version"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	Timestamp     Time              `json:"timestamp,omitempty"`
	IdleTimeLimit Duration          `json:"idle_time_limit,omitempty"`
	Environment   map[string]string `json:"env,omitempty"`
	Title         string            `json:"title,omitempty"`
}
