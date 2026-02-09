package idgen

import "github.com/sony/sonyflake"

var sf = sonyflake.NewSonyflake(sonyflake.Settings{
	MachineID: func() (uint16, error) {
		return 1, nil // 每个服务实例唯一
	},
})

func NextID() (int64, error) {
	id, err := sf.NextID()
	return int64(id), err
}
