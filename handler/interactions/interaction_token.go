package interactions

import (
	"encoding/json"
	"errors"
	"time"
)

func New(tkn string, createdAt time.Time) Token {
	return Token{
		token: &interactionToken{
			token:     tkn,
			createdAt: createdAt,
		},
	}
}

type token interface {
	json.Marshaler
	Get() (string, error)
	IsValid() bool
}

type interactionToken struct {
	token     string
	createdAt time.Time
}

func (t *interactionToken) UnmarshalJSON(buf []byte) error {
	v := struct {
		Token     string    `json:"token"`
		CreatedAt time.Time `json:"created_at"`
	}{}
	defer func() {
		t.token = v.Token
		t.createdAt = v.CreatedAt
	}()
	return json.Unmarshal(buf, &v)
}

func (t interactionToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Token     string    `json:"token"`
			CreatedAt time.Time `json:"created_at"`
		}{
			Token:     t.token,
			CreatedAt: t.createdAt,
		},
	)
}

// トークンを取得する
// 無効な場合エラーを返す
func (t interactionToken) Get() (string, error) {
	if t.createdAt.Compare(time.Now()) != -1 {
		return "", errors.New("error: expired token")
	}
	return t.token, nil
}

// トークンが有効か否か
func (t interactionToken) IsValid() bool {
	return t.createdAt.Compare(time.Now()) == -1
}
