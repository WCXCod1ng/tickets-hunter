package define

import "time"

// 订单过期时间常量
const ExpireDuration = 15 * 60 * time.Second // 订单过期时间，单位为秒，这里设置为15分钟
//const ExpireDuration = 1 * 60 * time.Second // 订单过期时间，单位为秒，这里设置为15分钟
