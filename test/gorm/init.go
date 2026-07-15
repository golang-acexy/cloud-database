package gorm

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-parent/parent"
)

var starterLoader *parent.StarterLoader

func init() {
	logger.EnableConsole(logger.TraceLevel, false)
	starterLoader = parent.InitStarterLoader([]parent.Starter{
		&gormstarter.GormStarter{
			Config: gormstarter.GormConfig{
				MySQL: &gormstarter.MySQLConfig{
					DatabaseConfig: gormstarter.DatabaseConfig{
				Username: "root",
				Password: "root",
				Database: "test",
				Host:     "127.0.0.1",
				Port:     13306,
					},
				},
			},
		},
	})
}
