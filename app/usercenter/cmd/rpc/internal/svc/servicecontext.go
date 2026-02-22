package svc

import (
	"tickets-hunter/app/usercenter/cmd/rpc/internal/config"
	"tickets-hunter/app/usercenter/model"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
	Snowflake *snowflake.Node
}

func createSnowflake(c config.Config) *snowflake.Node {
	// 解析配置文件中的时间字符串
	t, err := time.Parse("2006-01-02", c.Snowflake.StartTime)
	if err != nil {
		// 如果格式不对，建议直接 Panic，强制修正配置
		logx.Errorf("Snowflake StartTime parse error: %v, use default", err)
		panic("Invalid Snowflake StartTime format, use YYYY-MM-DD")
	}

	// 设置全局 Epoch (关键步骤)
	// 注意：snowflake.Epoch 是一个全局变量，必须在 NewNode 之前设置
	snowflake.Epoch = t.UnixMilli()

	// 初始化 snowflake 节点
	node, err := snowflake.NewNode(c.Snowflake.WorkerId)
	if err != nil {
		// 如果初始化失败（例如 WorkerId 超出范围），记录日志并 panic
		logx.Errorf("init snowflake node failed: %v", err)
		panic(err)
	}

	return node
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(mysqlConn, c.UserCache),
		Snowflake: createSnowflake(c),
	}
}
