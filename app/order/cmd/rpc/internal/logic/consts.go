package logic

import "time"

// 订单状态常量
const (
	OrderStatusPending = 10 // 新建/待支付
	//OrderStatusPaying    = 11 // 支付中
	OrderStatusPaidWaitIssue = 20 // 已支付/待出票，收到支付成功的回调后，此时钱已经从用户账户扣除，通知下游票务系统出票
	//OrderStatusIssuing       = 21 // 出票中，通知下游票务系统出票后，等待出票结果的过程中
	OrderStatusIssued          = 30 // 已出票/已完成，说明订单完成。在后续阶段中需要将座位状态改为已售出
	OrderStatusClosedByTimeout = 40 // 支付超时关闭，未支付成功的订单在一定时间后会被系统自动关闭，此时需要将座位状态改回可用
	//OrderStatusClosedByUser    = 41  // 用户主动取消订单，未支付成功的订单被用户取消，此时需要将座位状态改回可用
	//OrderStatusRefunding    = 50 // 退款中，用户申请退款后，等待客服审核的过程中
	OrderStatusRefunded = 51 // 已退款，客服审核通过后，完成退款，此时需要将座位状态改回可用
	//OrderStatusRefundDenied = 52 // 退款拒绝，客服审核拒绝后，订单状态改为退款拒绝，此时订单仍然有效，座位状态不变
)

const ExpireDuration = 15 * 60 * time.Second // 订单过期时间，单位为秒，这里设置为15分钟
