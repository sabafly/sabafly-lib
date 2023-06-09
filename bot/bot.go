/*
	Copyright (C) 2022-2023  sabafly

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package botlib

import (
	"context"
	"fmt"

	"github.com/sabafly/sabafly-lib/v2/handler"

	"github.com/disgoorg/log"
	"github.com/sabafly/disgo"
	"github.com/sabafly/disgo/bot"
	"github.com/sabafly/disgo/cache"
	"github.com/sabafly/disgo/discord"
	"github.com/sabafly/disgo/events"
	"github.com/sabafly/disgo/gateway"
	"github.com/sabafly/disgo/oauth2"
	"github.com/sabafly/disgo/sharding"
)

func New[T any](logger log.Logger, version string, config Config) *Bot[T] {
	return &Bot[T]{
		Logger:  logger,
		Config:  config,
		OAuth:   oauth2.New(config.ClientID, config.Secret, oauth2.WithLogger(logger)),
		Version: version,
		Handler: handler.New(logger),
	}
}

type Bot[T any] struct {
	Logger  log.Logger
	Client  bot.Client
	OAuth   oauth2.Client
	Config  Config
	Version string
	Handler *handler.Handler
	Self    T
}

func (b *Bot[T]) SetupBot(listeners ...bot.EventListener) {
	var err error
	b.Client, err = disgo.New(b.Config.Token,
		bot.WithLogger(b.Logger),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagsAll)),
		bot.WithShardManagerConfigOpts(sharding.WithAutoScaling(true), sharding.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentsAll), gateway.WithAutoReconnect(true), gateway.WithLogger(b.Logger))),
		bot.WithMemberChunkingFilter(bot.MemberChunkingFilterAll),
		bot.WithEventManagerConfigOpts(bot.WithAsyncEventsEnabled(), bot.WithListeners(listeners...), bot.WithEventManagerLogger(b.Logger)),
	)
	if err != nil {
		b.Logger.Fatalf("botのセットアップに失敗 %s", err)
	}
}

func (b *Bot[T]) OnGuildJoin(g *events.GuildJoin) {
	b.Logger.Infof("[#%d]ギルド参加 %3dメンバー 作成 %s 名前 %s(%d)", g.ShardID(), g.Guild.MemberCount, g.Guild.CreatedAt().String(), g.Guild.Name, g.GuildID)
	go b.RefreshPresence()
}

func (b *Bot[T]) OnGuildLeave(g *events.GuildLeave) {
	b.Logger.Infof("[#%d]ギルド脱退 %3dメンバー 作成 %s 名前 %s(%d)", g.ShardID(), g.Guild.MemberCount, g.Guild.CreatedAt().String(), g.Guild.Name, g.GuildID)
	b.Client.Caches().RemoveGuild(g.GuildID)
	b.Client.Caches().RemoveMembersByGuildID(g.GuildID)
	go b.RefreshPresence()
}

func (b *Bot[T]) OnGuildMemberJoin(m *events.GuildMemberJoin) {
	if g, ok := m.Client().Caches().Guild(m.GuildID); ok {
		b.Logger.Infof("[#%d]ギルドメンバー参加 %32s#%s(%d) ギルド %s(%d) %3d メンバー", m.ShardID(), m.Member.User.Username, m.Member.User.Discriminator, m.Member.User.ID, g.Name, g.ID, g.MemberCount)
	}
	go b.RefreshPresence()
}

func (b *Bot[T]) OnGuildMemberLeave(m *events.GuildMemberLeave) {
	if g, ok := m.Client().Caches().Guild(m.GuildID); ok {
		b.Logger.Infof("[#%d]ギルドメンバー脱退 %32s#%s(%d) ギルド %s(%d) %3d メンバー", m.ShardID(), m.Member.User.Username, m.Member.User.Discriminator, m.Member.User.ID, g.Name, g.ID, g.MemberCount)
	}
	b.Client.Caches().RemoveMember(m.GuildID, m.User.ID)
	go b.RefreshPresence()
}

func (b *Bot[T]) RefreshPresence() {
	var (
		guilds int = b.Client.Caches().GuildsLen()
		users  int = b.Client.Caches().MembersAllLen()
	)
	shards := b.Client.ShardManager().Shards()
	for k := range shards {
		state := fmt.Sprintf("/help | %d Servers | %d Users | #%d", guilds, users, k)
		if err := b.Client.SetPresenceForShard(context.TODO(), k, gateway.WithOnlineStatus(discord.OnlineStatusOnline), gateway.WithPlayingActivity(state)); err != nil {
			b.Logger.Errorf("ステータス更新に失敗 %s", err)
		}
	}
	if len(shards) == 0 {
		state := fmt.Sprintf("/help | %d Servers | %d Users", guilds, users)
		err := b.Client.SetPresence(context.TODO(), gateway.WithPlayingActivity(state))
		if err != nil {
			b.Logger.Errorf("ステータス更新に失敗 %s", err)
		}
	}
}
