package chat

import (
	"encoding/json"
	"fmt"
	"github.com/go-tubes/tubes"
)

type Chat struct {
	users      map[string]bool
	prefix     string
	tubeSystem *tubes.TubeSystem
}

func New(prefix string, tubeSystem *tubes.TubeSystem) *Chat {
	chat := &Chat{
		users:      map[string]bool{},
		prefix:     prefix,
		tubeSystem: tubeSystem,
	}

	tubeSystem.RegisterChannel(fmt.Sprintf("/%s/users", prefix), tubes.ChannelHandlers{
		OnSubscribe:   chat.onUserJoin,
		OnUnsubscribe: chat.onUserLeave,
	})

	tubeSystem.RegisterChannel(fmt.Sprintf("/%s", prefix), tubes.ChannelHandlers{
		OnMessage: chat.onChatMessage,
	})

	return chat
}

func (c *Chat) broadcastUsers(s *tubes.Context) {
	payload, _ := json.Marshal(c.users)
	s.Broadcast(payload, &tubes.ContextBroadcastOptions{
		ExcludeContextOwner: false,
	})
}

func (c *Chat) onChatMessage(s *tubes.Context, message *tubes.Message) {
	println("Received Message: " + s.FullPath)
	payload, _ := json.Marshal(fmt.Sprintf("%s: %s", s.Client.Id, message.Payload))
	s.Broadcast(payload, &tubes.ContextBroadcastOptions{
		ExcludeContextOwner: false,
	})
}

func (c *Chat) onUserJoin(s *tubes.Context) {
	println("Client joined: " + s.FullPath)
	c.users[s.Client.Id] = true
	c.broadcastUsers(s)
}

func (c *Chat) onUserLeave(s *tubes.Context) {
	println("Client left: " + s.FullPath)
	c.users[s.Client.Id] = false
	c.broadcastUsers(s)
}
