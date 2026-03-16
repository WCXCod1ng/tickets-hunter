package job

// OneShotJob 是一种特殊类型的 Job，它只会执行一次，执行完成后就不再调度。适用于那些只需要执行一次的任务，比如系统初始化、数据迁移等。
type OneShotJob interface {
	Job
	Once() bool
}
