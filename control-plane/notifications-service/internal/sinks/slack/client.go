package slack

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

type SlackClient struct {
	Slack *slack.Client
}

func NewSlackClient(url, apiToken string) *SlackClient {
	sc := slack.New(
		apiToken,
		slack.OptionAPIURL(url),
	)
	return &SlackClient{
		Slack: sc,
	}
}

func (c *SlackClient) AuthTest() (*slack.AuthTestResponse, error) {
	return c.Slack.AuthTest()
}

func (c *SlackClient) PostMessage(channel string, messageBlocks []slack.Block) error {
	options := []slack.MsgOption{
		slack.MsgOptionBlocks(messageBlocks...),
		slack.MsgOptionEnableLinkUnfurl(),
	}

	log.Debug().Msgf("Posting message to channel: %s \nblocks: %+v.\nwith options: %+v", channel, messageBlocks, options)
	_, _, err := c.Slack.PostMessage(channel, options...)
	if err != nil {
		return errors.Wrap(err, "could not post message")
	}
	return nil
}

func (c *SlackClient) PostDMMessage(userIds []string, messageBlocks []slack.Block) error {
	channel, _, _, err := c.Slack.OpenConversation(&slack.OpenConversationParameters{
		ReturnIM: true,
		Users:    userIds,
	})
	if err != nil {
		return errors.Wrap(err, "could not open DM channel")
	}
	return c.PostMessage(channel.ID, messageBlocks)
}

func (c *SlackClient) GetUserByEmail(email string) (*slack.User, error) {
	user, err := c.Slack.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
