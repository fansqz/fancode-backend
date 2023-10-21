//go:build wireinject
// +build wireinject

package main

import (
	"FanCode/config"
	"FanCode/controller"
	"FanCode/dao"
	"FanCode/interceptor"
	"FanCode/routers"
	"FanCode/service"
	"github.com/google/wire"
	"net/http"
)

func initApp(*config.AppConfig) (*http.Server, error) {
	panic(wire.Build(
		dao.ProviderSet,
		service.ProviderSet,
		controller.ProviderSet,
		interceptor.ProviderSet,
		routers.SetupRouter,
		newApp))
}
