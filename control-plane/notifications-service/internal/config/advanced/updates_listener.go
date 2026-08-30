package advanced

import (
	"context"

	"github.com/rs/zerolog/log"
)

type ConfigurationUpdatesListener struct {
	listenerUpdatesChannel chan string
}

func NewUpdatesListener(updatesChan chan string) *ConfigurationUpdatesListener {
	return &ConfigurationUpdatesListener{
		listenerUpdatesChannel: updatesChan,
	}
}

func (cul *ConfigurationUpdatesListener) Listen(ctx context.Context, handleListenerChange func(listenerId string) error) {
	go cul.listen(ctx, handleListenerChange)
}

func (cul *ConfigurationUpdatesListener) listen(ctx context.Context, handleListenerChange func(listenerId string) error) {
	for {
		select {
		case listenerId := <-cul.listenerUpdatesChannel:
			err := handleListenerChange(listenerId)
			if err != nil {
				log.Err(err).Msg("config change detected and encountered an error")
			}
		case <-ctx.Done():
			close(cul.listenerUpdatesChannel)
			return
		}
	}
}
