package interactions

import "encoding/json"

type Token struct {
	token
}

func (t *Token) UnmarshalJSON(buf []byte) error {
	v := interactionToken{}
	defer func() { t.token = v }()
	return json.Unmarshal(buf, &v)
}
