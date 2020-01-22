package streamer

type Message struct {
	ID      string      `json:"id"`
	Stream  string      `json:"stream"`
	Payload interface{} `json:"payload"`
}
