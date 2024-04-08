package proxmox

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (t *TokenInfo) UnmarshalJSON(data []byte) error {
	if t == nil {
		t = new(TokenInfo)
	}

	var rawInfo map[string]any
	err := json.Unmarshal(data, &rawInfo)
	if err != nil {
		return err
	}

	if rawExp, ok := rawInfo["expire"]; ok {
		switch e := rawExp.(type) {
		case string:
			if t.Expire, err = strconv.ParseInt(e, 10, 64); err != nil {
				return fmt.Errorf("could not parse expire time: %w", err)
			}
		case int:
			t.Expire = int64(e)
		case int32:
			t.Expire = int64(e)
		case uint:
			t.Expire = int64(e)
		case uint32:
			t.Expire = int64(e)
		case uint64:
			t.Expire = int64(e)
		case int64:
			t.Expire = e
		default:
			return fmt.Errorf("expected a string or int, got a %T", e)
		}
	}

	return nil
}

func (i *IntOrString) UnmarshalJSON(data []byte) error {
	s := string(data)
	s = strings.Trim(s, `"`)

	if i == nil {
		i = new(IntOrString)
	}

	iv, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	*i = IntOrString(iv)
	return nil
}
