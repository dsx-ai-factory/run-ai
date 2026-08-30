package queuesManager

type EventType int

const (
	TaskTimerIdle    EventType = iota
	TaskTimerTimeout EventType = iota
)

type ActionList struct {
	list         []int
	idleIndex    int
	timeoutIndex int
	index        *int
}

func NewActionList(list []int, idleIndex int, timeoutIndex int, index *int) *ActionList {
	return &ActionList{list: list, idleIndex: idleIndex, timeoutIndex: timeoutIndex, index: index}
}

func (actionList *ActionList) GetList() []int {
	return actionList.list
}

func (actionList *ActionList) GetListIndex() int {
	return actionList.list[*actionList.index]
}

func (actionList *ActionList) SetEventType(eventType EventType) *ActionList {
	switch eventType {
	case TaskTimerIdle:
		actionList.index = &actionList.idleIndex
	case TaskTimerTimeout:
		actionList.index = &actionList.timeoutIndex
	}
	return actionList
}

func (actionList *ActionList) SetIndex(index int) {
	*actionList.index = index
}

func (actionList *ActionList) GetIndex() int {
	return *actionList.index
}

func (actionList *ActionList) IncrementIndex() {
	*actionList.index++
}
