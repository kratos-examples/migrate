// migrate-kit is the sibling migration program of the two demo services. It runs
// hand-written go-migrate scripts against the same postgres database the services
// share, so the schema comes from scripted migration and not from gorm AutoMigrate.
//
// migrate-kit 是两个演示服务的旁挂迁移程序：针对服务共用的那个 postgres 库运行手写的
// go-migrate 脚本，让表结构来自脚本化迁移，而不是 gorm 的 AutoMigrate。
package main

import (
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/spf13/cobra"
	"github.com/yylego/go-migrate/cobramigration"
	"github.com/yylego/go-migrate/migrationparam"
	"github.com/yylego/go-migrate/newmigrate"
	"github.com/yylego/must"
	"github.com/yylego/rese"
	"github.com/yylego/runpath"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config maps the YAML config: the same postgres database the two demo services
// share, plus the path that holds the hand-written migration scripts.
//
// Config 对应 YAML 配置：两个演示服务共用的那个 postgres 库，外加存放手写迁移脚本的路径
type Config struct {
	Database struct {
		Driver string `yaml:"driver"`
		Source string `yaml:"source"`
	} `yaml:"database"`
	Scripts struct {
		Path string `yaml:"path"`
	} `yaml:"scripts"`
}

// loadConfig reads and parses the YAML config file.
//
// loadConfig 读取并解析 YAML 配置文件
func loadConfig(path string) *Config {
	data := rese.V1(os.ReadFile(path))
	var cfg Config
	must.Done(yaml.Unmarshal(data, &cfg))
	must.Nice(cfg.Database.Source)
	must.Nice(cfg.Scripts.Path)
	return &cfg
}

// resolveConfigPath picks the config file: --conf=xxx when passed, otherwise the
// config.yaml that sits next to this program.
//
// resolveConfigPath 选择配置文件：传了 --conf=xxx 就读你的，否则读本程序旁边的 config.yaml
func resolveConfigPath() string {
	for i, arg := range os.Args {
		if arg == "--conf" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
		if strings.HasPrefix(arg, "--conf=") {
			return strings.TrimPrefix(arg, "--conf=")
		}
	}
	return runpath.PARENT.Join("config.yaml")
}

func main() {
	cfg := loadConfig(resolveConfigPath())
	scriptsPath := cfg.Scripts.Path

	param := migrationparam.NewMigrationParam(
		func() *gorm.DB {
			must.Same(cfg.Database.Driver, "postgres")
			return rese.P1(gorm.Open(postgres.Open(cfg.Database.Source), &gorm.Config{}))
		},
		func(db *gorm.DB) *migrate.Migrate {
			rawDB := rese.P1(db.DB())
			driver := rese.V1(postgresmigrate.WithInstance(rawDB, &postgresmigrate.Config{}))
			return rese.P1(newmigrate.NewWithScriptsAndDatabase(
				&newmigrate.ScriptsAndDatabaseParam{
					ScriptsInRoot:    scriptsPath,
					DatabaseName:     "postgres",
					DatabaseInstance: driver,
				},
			))
		},
	)

	var debugMode bool

	rootCmd := &cobra.Command{
		Use:   "migrate-kit",
		Short: "Scripted database migration on the shared demo database",
		Args:  cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			migrationparam.SetDebugMode(debugMode)
		},
	}
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().String("conf", "", "config file path (default: the config.yaml next to this program)")

	// migrate all: run every pending step forward
	// migrate inc: step one forward
	// migrate dec: step one back
	//
	// migrate all：向前跑完所有待迁移步骤
	// migrate inc：向前走一步
	// migrate dec：向后退一步
	rootCmd.AddCommand(cobramigration.NewMigrateCmd(param))

	must.Done(rootCmd.Execute())
}
