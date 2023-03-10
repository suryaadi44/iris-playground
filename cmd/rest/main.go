package main

import (
	"context"
	"log"
	"net"

	"github.com/suryaadi44/iris-playground/app/api/rest"
	"github.com/suryaadi44/iris-playground/utils/config"
	"github.com/suryaadi44/iris-playground/utils/database/postgresql"
	"github.com/suryaadi44/iris-playground/utils/shutdown"

	"time"

	"github.com/kataras/iris/v12"
)

func main() {
	config, err := config.Load("config")
	if err != nil {
		log.Fatalf("[Config] Error loading config: %v", err)
	}

	local, err := time.LoadLocation(config.GetString("timezone"))
	if err != nil {
		log.Fatalf("[Config] Error loading timezone: %v", err)
	}
	time.Local = local

	db := postgresql.InitDatabase(
		config.GetString("database.postgres.host"),
		config.GetInt("database.postgres.port"),
		config.GetString("database.postgres.user"),
		config.GetString("database.postgres.password"),
		config.GetString("database.postgres.database"),
		config.GetString("timezone"),
	)

	app := iris.New()

	rest.InitRoute(app, db, config)

	listen := net.JoinHostPort(config.GetString("rest.host"), config.GetString("rest.port"))
	go func() {
		if err := app.Listen(listen, iris.WithoutInterruptHandler, iris.WithoutServerError(iris.ErrServerClosed)); err != nil {
			log.Fatalf("[Server] Error running server: %v", err)
		}
	}()

	ctx := context.Background()
	// operations to be executed on shutdown
	ops := map[string]shutdown.Operation{
		// TODO: Add more operations here
		// "redis": ...

		"database": func(ctx context.Context) error {
			dbCon, err := db.DB()
			if err != nil {
				return err
			}
			return dbCon.Close()
		},
		"rest-server": func(ctx context.Context) error {
			return app.Shutdown(ctx)
		},
	}
	// listen for interrupt signal
	<-shutdown.GracefulShutdown(ctx, config.GetDuration("shutdown_timeout"), ops)
}
