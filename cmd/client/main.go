package main

import (
	"github.com/g123udini/gophkeeper/internal/client/app"
	"github.com/g123udini/gophkeeper/internal/common/buildlog"
	"github.com/g123udini/gophkeeper/internal/common/logger"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	buildlog.Print(buildVersion, buildDate, buildCommit)
	logger.Init("client", zap.InfoLevel.String())

	app, err := app.New()
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	app.Run()
}
