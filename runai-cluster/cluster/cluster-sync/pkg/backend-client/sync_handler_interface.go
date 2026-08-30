package backendclient

import (
	"sync"
)

type SyncHandler interface {
	HandleSyncEvent(waitGroup *sync.WaitGroup)
}
