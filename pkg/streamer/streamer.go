package streamer

type Message struct {
	ID      string      `json:"id"`
	Payload interface{} `json:"payload"`
}
