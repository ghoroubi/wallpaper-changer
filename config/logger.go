package config

import (
	"context"
	"time"

	"github.com/ghoroubi/gerrors"
	"github.com/mattn/go-colorable"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"gopkg.in/sohlich/elogrus.v7"

	"bedsonline/config/env"
	rotate "bedsonline/helpers/file_rotator"
	"bedsonline/models"
)

// GetLogger ...
// Initializes and returns the logger instance
func GetLogger() *logrus.Logger {

	var (
		// Setting log level
		logLevel = logrus.Level(uint32(env.GetInt("services.log.logrus.log_level")))

		// Instantiation the logger
		l = logrus.StandardLogger()

		// Service debug mode
		debug = env.GetBool("debug")
	)

	// Check for debug level
	if debug {
		logLevel = logrus.DebugLevel
	}

	// File rotator init
	// Setting up a file as logger with custom attributes which comes from config file
	rotateFileHook, err := rotate.NewRotateFileHook(rotate.RotateFileConfig{
		Filename:   env.GetString("services.log.logrus.log_file"),
		MaxSize:    env.GetInt("services.log.logrus.max_size"),
		MaxBackups: env.GetInt("services.log.logrus.max_backup"),
		MaxAge:     env.GetInt("services.log.logrus.max_age"),
		Level:      logLevel,
		Formatter: &logrus.JSONFormatter{
			DisableHTMLEscape: true,
			PrettyPrint:       env.GetBool("services.log.logrus.pretty_format"),
		},
	})
	if err != nil {
		logrus.Fatalf("Failed to initialize file rotate hook: %v", err)
	}

	// RotateFileHook for logrus
	// Which streams the log data into the provided file hook
	l.AddHook(rotateFileHook)

	// Set level for main logger instance
	l.SetLevel(logLevel)

	// Set report caller which reports the caller function of logger operation
	l.SetReportCaller(false)

	// Print to StdOut only in debug mode
	if debug {
		l.SetOutput(colorable.NewColorableStdout())
	}

	// Stream to elastic and logstash
	// only in production mode
	if !debug {
		// Adding logstash hook which streams the log data to the logstash engine
		logstashHook := getLogstashHook("tcp",
			env.GetString("services.log.logrus.logstash.address"),
			env.GetString("name"))
		if logstashHook != nil {
			l.AddHook(logstashHook)
		}

		// Adding elastic hook which streams the log data to the elastic dataset
		elasticHook, err := getElasticHook(env.GetString("services.log.logrus.elastic.address"),
			env.GetString("name"))
		if err != nil {
			logrus.Warningf("Failed to initialize elastic hook: %v \n", err)
			return l
		}

		// Adding elastic hook
		l.AddHook(elasticHook)
	}

	return l
}

// Logstash hook for logrus
func getLogstashHook(network string, addr string, logo string) logrus.Hook {
	/*if network == "" {
		network = "tcp"
	}
	hook, err := logrustash.NewHook(network, addr, logo)
	if err != nil {
		return nil
	}
	return hook*/
	return nil
}

// ElasticSearch hook for logrus
func getElasticHook(addr string, svcName string) (logrus.Hook, error) {

	// Validate elastic address
	if addr == "" {
		return nil, gerrors.WrapError(models.ErrInternalServerError, models.ErrElasticNodeNotFound)
	}

	// New elastic client
	d, err := time.ParseDuration(env.GetString("services.log.logrus.elastic.connection_time_out"))
	if err != nil {
		return nil, gerrors.WrapError(models.ErrInternalServerError, err)
	}

	// The timeout of elastic connection is too long,
	// Thus a context with timeout is required which
	// Gets the amount from config file
	ctx, cancelFunc := context.WithTimeout(context.Background(), d)

	// Creating elastic client with above context
	client, err := elastic.DialContext(ctx, elastic.SetURL(addr))
	if err != nil {
		return nil, gerrors.WrapError(models.ErrInternalServerError, err)
	}

	// cancel context as soon as deadline occurs
	defer cancelFunc()

	// New elastic async hook
	hook, err := elogrus.NewAsyncElasticHook(client, "localhost", logrus.DebugLevel, svcName)
	if err != nil {
		return nil, gerrors.WrapError(models.ErrInternalServerError, err)
	}

	return hook, nil
}
