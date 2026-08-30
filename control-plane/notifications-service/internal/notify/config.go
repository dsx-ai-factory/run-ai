package notify

type NotifyingConfigProvider interface {
	GetNotifyIds() []string
}
