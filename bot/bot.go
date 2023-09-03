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
	"github.com/sabafly/sabafly-lib/v2/handler"

	"github.com/disgoorg/log"
	"github.com/sabafly/disgo"
	"github.com/sabafly/disgo/bot"
	"github.com/sabafly/disgo/cache"
	"github.com/sabafly/disgo/gateway"
	"github.com/sabafly/disgo/handlers"
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
		bot.WithEventManagerConfigOpts(bot.WithAsyncEventsEnabled(), bot.WithListeners(listeners...), bot.WithEventManagerLogger(b.Logger), bot.WithGatewayHandlers(handlers.GetGatewayHandlers())),
	)
	if err != nil {
		b.Logger.Fatalf("botのセットアップに失敗 %s", err)
	}
}
