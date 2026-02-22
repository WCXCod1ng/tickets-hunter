package entity

type RefreshInfo struct {
	UserId   string `json:"userId" msgpack:"userId"`
	Platform string `json:"platform" msgpack:"platform"`
}
