package serialize

import "github.com/vmihailenco/msgpack/v5"

type MsgPackSerializer struct{}

func (MsgPackSerializer) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (MsgPackSerializer) Unmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

func (MsgPackSerializer) Name() string {
	return "msgpack"
}
