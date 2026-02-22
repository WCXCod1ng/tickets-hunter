package serialize

import "encoding/json"

type JSONSerializer struct{}

func (JSONSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (JSONSerializer) Name() string {
	return "json"
}
