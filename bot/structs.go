package botlib

import "github.com/disgoorg/snowflake/v2"

type MessageLog struct {
	ID        snowflake.ID
	GuildID   snowflake.ID
	ChannelID snowflake.ID
	UserID    snowflake.ID
	Content   string
	Bot       bool
}
