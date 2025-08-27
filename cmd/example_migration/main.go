package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"

	"github.com/razorpay/goutils/configloader"
	logger "github.com/razorpay/goutils/logger/v3"
	storesql "github.com/razorpay/goutils/sqlstorage/sql"

	cfg "github.com/razorpay/go-foundation-v2/cmd/example/config"
	_ "github.com/razorpay/go-foundation-v2/internal/example/migrations"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String(
		"dir",
		"internal/example/payment/migrations",
		"Directory with migration files",
	)
	verbose = flags.Bool("v", false, "Enable verbose mode")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slogger := slog.New(logger.NewHandler(nil)).
		With(slog.String("service", "example-migration"))

	// load configurations and distribute parts of it in main
	config := cfg.Config{}
	env := configloader.GetAppEnv()
	loader := configloader.New(
		configloader.WithConfigDir("./config/example"),
	)
	if err := loader.Load(env, &config); err != nil {
		logger.WithError(slogger, err).Error("could not load config")
		os.Exit(1)
	}

	err := goose.SetDialect(config.Store.SQL.Dialect.String())
	if err != nil {
		logger.WithError(slogger, err).Error("could not set dialect")
		os.Exit(1)
	}

	// storage service is the service for main persistent store
	gorm, err := storesql.NewGorm(ctx, &config.Store.SQL)
	if err != nil {
		logger.WithError(slogger, err).Error("could not create gorm")
		os.Exit(1)
	}

	sqlDB, err := gorm.DB(ctx)
	if err != nil {
		logger.WithError(slogger, err).Error("could not get the dbinstance")
		os.Exit(1)
	}

	run(ctx, sqlDB, *dir, slogger)
}

func run(ctx context.Context, db *sql.DB, dir string, slogger *slog.Logger) {
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		logger.WithError(slogger, err).Error("error parsing flags")
		os.Exit(1)
	}
	args := flags.Args()
	if *verbose {
		goose.SetVerbose(true)
	}

	// I.e. no command provided, hence print usage and return.
	if len(args) < 1 {
		cmd := os.Getenv("MIGRATION_CMD")
		if cmd == "" {
			flags.Usage()
			return
		}

		args = append(args, cmd)

	}
	// Prepares command and arguments for goose's run.
	command := args[0]
	arguments := []string{}
	if len(args) > 1 {
		arguments = append(arguments, args[1:]...)
	}

	// If command is create or fix, no need to connect to db and hence the
	// specific case handling.
	switch command {
	case "create":
		fmt.Println(dir)
		err := goose.RunContext(ctx, "create", nil, dir, arguments...)
		if err != nil {
			logger.WithError(slogger, err).Error("error parsing flags")
			os.Exit(1)
		}
		return
	case "fix":
		err := goose.RunContext(ctx, "fix", nil, dir)
		if err != nil {
			logger.WithError(slogger, err).Error("error parsing flags")
			os.Exit(1)
		}
		return
	}

	// Finally, executes the goose's command.
	err := goose.RunContext(ctx, command, db, dir, arguments...)
	if err != nil {
		logger.WithError(slogger, err).Error("error parsing flags")
		os.Exit(1)
	}

}

func usage() {
	flags.PrintDefaults()
	fmt.Println(usageCommands)
}

var usageCommands = `
	Commands:
		up                   Migrate the DB to the most recent version available
		up-to VERSION        Migrate the DB to a specific VERSION
		down                 Roll back the version by 1
		down-to VERSION      Roll back to a specific VERSION
		redo                 Re-run the latest migration
		reset                Roll back all migrations
		status               Dump the migration status for the current DB
		version              Print the current version of the database
		create NAME          Creates new migration file with the current timestamp
		fix                  Apply sequential ordering to migrations
`
