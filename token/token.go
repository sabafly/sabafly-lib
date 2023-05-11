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
package token

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

var (
	ErrInvalidTokenFormat = errors.New("invalid token format")
)

func MustParse(str string) Token {
	t, err := Parse(str)
	if err != nil {
		panic(err)
	}
	return t
}

func Parse(str string) (Token, error) {
	part := strings.Split(str, ".")
	if len(part) != 3 {
		return Token{}, ErrInvalidTokenFormat
	}
	decode_str, err := base64.RawStdEncoding.DecodeString(part[0])
	if err != nil {
		return Token{}, err
	}
	id, err := snowflake.Parse(string(decode_str))
	if err != nil {
		return Token{}, err
	}
	if decode_str, err = base64.RawStdEncoding.DecodeString(part[1]); err != nil {
		return Token{}, err
	}
	time_int := new(big.Int).SetBytes([]byte(decode_str)).Int64()
	timestamp := time.Unix(time_int+Epoch, 0)
	return Token{
		rand:      []byte(part[2]),
		timestamp: timestamp,
		id:        id,
	}, nil
}

func NewToken(id snowflake.ID) Token {
	return Token{
		rand:      genRandBytes(),
		timestamp: time.Now(),
		id:        id,
	}
}

func (t Token) Check(t2 Token) bool {
	return t.String() == t2.String()
}

type Token struct {
	rand      []byte
	timestamp time.Time
	id        snowflake.ID
}

const Epoch = 1293840000

func (t Token) String() string {
	return fmt.Sprintf("%s.%s.%s", base64.RawStdEncoding.EncodeToString([]byte(t.id.String())), base64.RawStdEncoding.EncodeToString(big.NewInt(time.Now().Unix()-Epoch).Bytes()), string(t.rand))
}

func (t Token) CreatedAt() time.Time {
	return t.timestamp
}

func (t Token) ID() snowflake.ID {
	return t.id
}

func (t Token) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.String() + "\""), nil
}

func (t *Token) UnmarshalJSON(b []byte) (err error) {
	str := string(b)
	str, _ = strings.CutPrefix(str, "\"")
	str, _ = strings.CutSuffix(str, "\"")
	*t, err = Parse(str)
	if err != nil {
		return err
	}
	return nil
}

var _ json.Marshaler = (*Token)(nil)
var _ json.Unmarshaler = (*Token)(nil)

var letters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_")

func genRandBytes() []byte {
	v := []byte{}
	for i := 0; i < 27; i++ {
		v = append(v, byte(letters[rand.Intn(len(letters)-1)]))
	}
	return v
}
