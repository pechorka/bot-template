package bot

type Echo struct {
}

func (e *Echo) OnMessage(message Message) Response {
	return Response{
		Text: message.Text,
		Send: true,
	}
}
