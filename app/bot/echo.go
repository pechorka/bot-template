package bot

// Echo is a simple bot that just repeats the message it receives.
type Echo struct {
}

// OnMessage will reply with the same message
func (e *Echo) OnMessage(message Message) Response {
	responseText := message.Text
	if responseText == "ping" {
		responseText = "pong"
	}
	return Response{
		Text: responseText,
		Send: true,
	}
}
