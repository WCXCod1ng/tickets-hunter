package svc

import (
	"tickets-hunter/app/model/ticket_event"
	"tickets-hunter/app/model/ticket_seat"
	"tickets-hunter/app/ticket/cmd/rpc/internal/config"
	"tickets-hunter/common/luaexec"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config           config.Config
	TicketEventModel ticket_event.TicketEventModel
	TicketSeatModel  ticket_seat.TicketSeatModel
	// Redis
	Redis *redis.Redis

	// 锁座Lua脚本
	LockSeatLuaScript *luaexec.LuaScript
	// 解锁座Lua脚本
	UnlockSeatLuaScript *luaexec.LuaScript
	// 兜底的解锁Lua脚本
	UnderwriteUnlockSeatScript *luaexec.LuaScript
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:                     c,
		TicketEventModel:           ticket_event.NewTicketEventModel(mysqlConn),
		TicketSeatModel:            ticket_seat.NewTicketSeatModel(mysqlConn),
		Redis:                      redis.MustNewRedis(c.Redis.RedisConf),
		LockSeatLuaScript:          luaexec.NewLuaScript(luaexec.MustLoadLuaFile("internal/lua/lockSeatScript.lua")),
		UnlockSeatLuaScript:        luaexec.NewLuaScript(luaexec.MustLoadLuaFile("internal/lua/unlockSeatScript.lua")),
		UnderwriteUnlockSeatScript: luaexec.NewLuaScript(luaexec.MustLoadLuaFile("internal/lua/underwriteUnlockSeatScript.lua")),
	}
}
