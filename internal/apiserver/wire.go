//go:build wireinject
// +build wireinject

package apiserver

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"github.com/ashwinyue/eino-show/internal/apiserver/biz"
	"github.com/ashwinyue/eino-show/internal/apiserver/cache"
	"github.com/ashwinyue/eino-show/internal/apiserver/pkg/validation"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	ginmw "github.com/ashwinyue/eino-show/internal/pkg/middleware/gin"
	"github.com/ashwinyue/eino-show/internal/pkg/server"
)

func InitializeWebServer(*Config) (server.Server, error) {
	wire.Build(
		wire.NewSet(NewWebServer, wire.FieldsOf(new(*Config), "ServerMode")),
		wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
		wire.NewSet(store.ProviderSet, biz.ProviderSet),
		ProvideDB, // 提供数据库实例
		wire.NewSet(
			ProvideRedisOrNone,
			wire.Bind(new(redis.UniversalClient), new(*redis.Client)),
			cache.NewRedisCache,
		),
		validation.ProviderSet,
		wire.NewSet(
			wire.Struct(new(UserRetriever), "*"),
			wire.Bind(new(ginmw.UserRetriever), new(*UserRetriever)),
		),
	)
	return nil, nil
}
