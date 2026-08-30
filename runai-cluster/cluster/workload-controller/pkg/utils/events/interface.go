package events

type Recorder interface {
	Event(eventtype, reason, message string)
	Eventf(eventtype, reason, messageFmt string, args ...interface{})
}
