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
package db

import (
	"fmt"

	"github.com/go-redis/redis/v8"
)

type DBConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	DB   int    `json:"db"`
}

type DB interface {
	Close() error
}

type DBRecord[T any, ID any] interface {
	Get(ID) (T, error)
	Set(ID, T) error
	Del(ID) error
}

func SetupDatabase(cfg DBConfig) (DB, error) {
	db := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		DB:      cfg.DB,
	})
	return &dbImpl{
		db: db,
	}, nil
}

type dbImpl struct {
	db *redis.Client
}

func (d *dbImpl) Close() error {
	return d.db.Close()
}
