package providers

type NotificationProvider interface {
	Name() string

	Init() error
	SendMessage(message string) error
}

func AddPrefix(prefix string, message string) string {
	if prefix == "" {
		return message
	}

	return prefix + ": " + message
}
