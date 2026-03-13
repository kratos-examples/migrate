# Changes

Code differences compared to source project.

## cmd/demo1kratos/cfgpath/cfg_path.go (+5 -0)

```diff
@@ -0,0 +1,5 @@
+package cfgpath
+
+// ConfigPath is the config path.
+// 配置文件路径
+var ConfigPath string
```

## cmd/demo1kratos/main.go (+33 -16)

```diff
@@ -1,19 +1,22 @@
 package main
 
 import (
-	"flag"
+	"fmt"
 	"os"
 
 	"github.com/go-kratos/kratos/v2"
-	"github.com/go-kratos/kratos/v2/config"
-	"github.com/go-kratos/kratos/v2/config/file"
 	"github.com/go-kratos/kratos/v2/log"
 	"github.com/go-kratos/kratos/v2/middleware/tracing"
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	"github.com/go-kratos/kratos/v2/transport/http"
+	"github.com/spf13/cobra"
 	"github.com/yylego/done"
+	"github.com/yylego/kratos-examples/demo1kratos/cmd/demo1kratos/cfgpath"
+	"github.com/yylego/kratos-examples/demo1kratos/cmd/demo1kratos/subcmds"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/appcfg"
 	"github.com/yylego/must"
+	"github.com/yylego/must/mustslice"
 	"github.com/yylego/rese"
 )
 
@@ -23,12 +26,10 @@
 	Name string
 	// Version is the version of the compiled software.
 	Version string
-	// flagconf is the config flag.
-	flagconf string
 )
 
 func init() {
-	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
+	fmt.Println("service-name:", Name)
 }
 
 func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
@@ -46,7 +47,6 @@
 }
 
 func main() {
-	flag.Parse()
 	logger := log.With(log.NewStdLogger(os.Stdout),
 		"ts", log.DefaultTimestamp,
 		"caller", log.DefaultCaller,
@@ -56,18 +56,35 @@
 		"trace.id", tracing.TraceID(),
 		"span.id", tracing.SpanID(),
 	)
-	c := config.New(
-		config.WithSource(
-			file.NewSource(flagconf),
-		),
-	)
-	defer rese.F0(c.Close)
 
-	must.Done(c.Load())
+	var rootCmd = &cobra.Command{
+		Use:   "demo1kratos",
+		Short: "A Kratos microservice with database migration",
+		Run: func(cmd *cobra.Command, args []string) {
+			mustslice.None(args)
+			if cfg := appcfg.ParseConfig(cfgpath.ConfigPath); cfg.Server.AutoRun {
+				runApp(cfg, logger)
+			}
+		},
+	}
+	rootCmd.PersistentFlags().StringVarP(&cfgpath.ConfigPath, "conf", "c", "./configs", "config path, eg: --conf=config.yaml")
 
-	var cfg conf.Bootstrap
-	must.Done(c.Scan(&cfg))
+	rootCmd.AddCommand(&cobra.Command{
+		Use:   "run",
+		Short: "Start the application",
+		Run: func(cmd *cobra.Command, args []string) {
+			cfg := appcfg.ParseConfig(cfgpath.ConfigPath)
+			runApp(cfg, logger)
+		},
+	})
 
+	rootCmd.AddCommand(subcmds.NewVersionCmd(Name, Version, logger))
+	rootCmd.AddCommand(subcmds.NewMigrateCmd(logger))
+
+	must.Done(rootCmd.Execute())
+}
+
+func runApp(cfg *conf.Bootstrap, logger log.Logger) {
 	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
 	defer cleanup()
 
```

## cmd/demo1kratos/subcmds/sub_cmds.go (+118 -0)

```diff
@@ -0,0 +1,118 @@
+package subcmds
+
+import (
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/golang-migrate/migrate/v4"
+	sqlite3migrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
+	"github.com/spf13/cobra"
+	"github.com/yylego/go-migrate/cobramigration"
+	"github.com/yylego/go-migrate/migrationparam"
+	"github.com/yylego/go-migrate/migrationstate"
+	"github.com/yylego/go-migrate/newmigrate"
+	"github.com/yylego/go-migrate/newscripts"
+	"github.com/yylego/go-migrate/previewmigrate"
+	"github.com/yylego/kratos-examples/demo1kratos/cmd/demo1kratos/cfgpath"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/appcfg"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
+	"github.com/yylego/must"
+	"github.com/yylego/rese"
+	"gorm.io/driver/sqlite"
+	"gorm.io/gorm"
+)
+
+// NewVersionCmd creates version command
+// 创建版本命令
+func NewVersionCmd(serviceName, version string, logger log.Logger) *cobra.Command {
+	return &cobra.Command{
+		Use:   "version",
+		Short: "Print version info",
+		Run: func(cmd *cobra.Command, args []string) {
+			slog := log.NewHelper(logger)
+			slog.Infof("service-name: %s version: %s", serviceName, version)
+		},
+	}
+}
+
+// NewMigrateCmd creates migrate command with database access
+// 创建带数据库访问的 migrate 命令
+//
+// Example commands:
+// 示例命令:
+//
+// Create migration scripts:
+// 创建迁移脚本:
+// ./bin/demo1kratos migrate new-script create --version-type TIME --description create_table
+// ./bin/demo1kratos migrate new-script create --version-type TIME --description alter_schema
+// ./bin/demo1kratos migrate new-script create --version-type TIME --description alter_schema --allow-empty-script true
+// ./bin/demo1kratos migrate new-script create --version-type TIME --description alter_column
+//
+// Update migration scripts:
+// 更新迁移脚本:
+// ./bin/demo1kratos migrate new-script update
+//
+// Execute migrations:
+// 执行迁移:
+// ./bin/demo1kratos migrate migrate all
+// ./bin/demo1kratos migrate migrate inc
+//
+// Preview migrations:
+// 预览迁移:
+// ./bin/demo1kratos migrate preview inc
+//
+// Check migration status:
+// 检查迁移状态:
+// ./bin/demo1kratos migrate status
+//
+// Note: Use caution with rollback operations to avoid unintended destructive actions
+// 注意: 回退操作要谨慎，避免误操作导致问题
+// ./bin/demo1kratos migrate migrate dec (use with caution)
+func NewMigrateCmd(logger log.Logger) *cobra.Command {
+	var debugMode bool
+
+	var rootCmd = &cobra.Command{
+		Use:   "migrate",
+		Short: "migrate",
+		Long:  "migrate",
+		Args:  cobra.NoArgs,
+		PersistentPreRun: func(cmd *cobra.Command, args []string) {
+			migrationparam.SetDebugMode(debugMode)
+		},
+	}
+	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug mode")
+
+	const scriptsInRoot = "./scripts"
+
+	param := migrationparam.NewMigrationParam(
+		func() *gorm.DB {
+			cfg := appcfg.ParseConfig(cfgpath.ConfigPath)
+			dsn := must.Nice(cfg.Data.Database.Source)
+			db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{}))
+			return db
+		},
+		func(db *gorm.DB) *migrate.Migrate {
+			rawDB := rese.P1(db.DB())
+			migrationDriver := rese.V1(sqlite3migrate.WithInstance(rawDB, &sqlite3migrate.Config{}))
+			return rese.P1(newmigrate.NewWithScriptsAndDatabase(
+				&newmigrate.ScriptsAndDatabaseParam{
+					ScriptsInRoot:    scriptsInRoot,
+					DatabaseName:     "sqlite3",
+					DatabaseInstance: migrationDriver,
+				},
+			))
+		},
+	)
+	rootCmd.AddCommand(newscripts.NewScriptCmd(&newscripts.Config{
+		Param:   param,
+		Options: newscripts.NewOptions(scriptsInRoot),
+		Objects: models.Objects(),
+	}))
+	rootCmd.AddCommand(cobramigration.NewMigrateCmd(param))
+	rootCmd.AddCommand(previewmigrate.NewPreviewCmd(param, scriptsInRoot))
+	rootCmd.AddCommand(migrationstate.NewStatusCmd(&migrationstate.Config{
+		Param:       param,
+		ScriptsPath: scriptsInRoot,
+		Objects:     models.Objects(),
+	}))
+
+	return rootCmd
+}
```

## configs/config.yaml (+3 -2)

```diff
@@ -5,7 +5,8 @@
   grpc:
     address: 0.0.0.0:9001
     timeout: 1s
+  auto_run: true
 data:
   database:
-    driver: sqlite3
-    source: file:db-C60AA5A0-DC6A-4E99-8583-2F9C2EEFF7A3?mode=memory&cache=shared
+    driver: sqlite
+    source: ./bin/demo1kratos.db
```

## internal/biz/student.go (+36 -4)

```diff
@@ -4,10 +4,13 @@
 	"context"
 
 	"github.com/brianvoe/gofakeit/v7"
+	"github.com/go-kratos/kratos/v2/errors"
 	"github.com/go-kratos/kratos/v2/log"
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
+	"gorm.io/gorm"
 )
 
 type Student struct {
@@ -27,10 +30,30 @@
 }
 
 func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
+	db := uc.data.DB()
+
+	// Use GORM transaction to save student
+	// 使用 GORM 事务保存学生
+	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
+		student := &models.Student{
+			Name: s.Name,
+		}
+		if err := tx.Create(student).Error; err != nil {
+			return err
+		}
+		s.ID = int64(student.ID)
+		return nil
+	})
+	if err != nil {
+		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("db: %v", err))
+	}
+
 	var res Student
 	if err := gofakeit.Struct(&res); err != nil {
 		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("fake: %v", err))
 	}
+	res.ID = s.ID
+	res.Name = s.Name
 	return &res, nil
 }
 
@@ -47,11 +70,20 @@
 }
 
 func (uc *StudentUsecase) GetStudent(ctx context.Context, id int64) (*Student, *ebzkratos.Ebz) {
-	var res Student
-	if err := gofakeit.Struct(&res); err != nil {
-		return nil, ebzkratos.New(pb.ErrorServerError("fake: %v", err))
+	db := uc.data.DB()
+
+	var student models.Student
+	if err := db.WithContext(ctx).First(&student, id).Error; err != nil {
+		if errors.Is(err, gorm.ErrRecordNotFound) {
+			return nil, ebzkratos.New(pb.ErrorServerError("not found: %v", err))
+		}
+		return nil, ebzkratos.New(pb.ErrorServerError("db: %v", err))
 	}
-	return &res, nil
+
+	return &Student{
+		ID:   int64(student.ID),
+		Name: student.Name,
+	}, nil
 }
 
 func (uc *StudentUsecase) ListStudents(ctx context.Context, page int32, pageSize int32) ([]*Student, int32, *ebzkratos.Ebz) {
```

## internal/conf/conf.pb.go (+11 -2)

```diff
@@ -78,6 +78,7 @@
 	state         protoimpl.MessageState `protogen:"open.v1"`
 	Http          *Server_HTTP           `protobuf:"bytes,1,opt,name=http,proto3" json:"http,omitempty"`
 	Grpc          *Server_GRPC           `protobuf:"bytes,2,opt,name=grpc,proto3" json:"grpc,omitempty"`
+	AutoRun       bool                   `protobuf:"varint,3,opt,name=auto_run,json=autoRun,proto3" json:"auto_run,omitempty"`
 	unknownFields protoimpl.UnknownFields
 	sizeCache     protoimpl.SizeCache
 }
@@ -126,6 +127,13 @@
 	return nil
 }
 
+func (x *Server) GetAutoRun() bool {
+	if x != nil {
+		return x.AutoRun
+	}
+	return false
+}
+
 type Data struct {
 	state         protoimpl.MessageState `protogen:"open.v1"`
 	Database      *Data_Database         `protobuf:"bytes,1,opt,name=database,proto3" json:"database,omitempty"`
@@ -350,10 +358,11 @@
 	"kratos.api\x1a\x1egoogle/protobuf/duration.proto\"]\n" +
 	"\tBootstrap\x12*\n" +
 	"\x06server\x18\x01 \x01(\v2\x12.kratos.api.ServerR\x06server\x12$\n" +
-	"\x04data\x18\x02 \x01(\v2\x10.kratos.api.DataR\x04data\"\xc4\x02\n" +
+	"\x04data\x18\x02 \x01(\v2\x10.kratos.api.DataR\x04data\"\xdf\x02\n" +
 	"\x06Server\x12+\n" +
 	"\x04http\x18\x01 \x01(\v2\x17.kratos.api.Server.HTTPR\x04http\x12+\n" +
-	"\x04grpc\x18\x02 \x01(\v2\x17.kratos.api.Server.GRPCR\x04grpc\x1ao\n" +
+	"\x04grpc\x18\x02 \x01(\v2\x17.kratos.api.Server.GRPCR\x04grpc\x12\x19\n" +
+	"\bauto_run\x18\x03 \x01(\bR\aautoRun\x1ao\n" +
 	"\x04HTTP\x12\x18\n" +
 	"\anetwork\x18\x01 \x01(\tR\anetwork\x12\x18\n" +
 	"\aaddress\x18\x02 \x01(\tR\aaddress\x123\n" +
```

## internal/conf/conf.proto (+1 -0)

```diff
@@ -23,6 +23,7 @@
   }
   HTTP http = 1;
   GRPC grpc = 2;
+  bool auto_run = 3;
 }
 
 message Data {
```

## internal/data/data.go (+25 -4)

```diff
@@ -3,25 +3,46 @@
 import (
 	"github.com/go-kratos/kratos/v2/log"
 	"github.com/google/wire"
+	"github.com/yylego/go-migrate/checkmigration"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
 	"gorm.io/driver/sqlite"
 	"gorm.io/gorm"
+	loggergorm "gorm.io/gorm/logger"
 )
 
+// ProviderSet is data providers.
 var ProviderSet = wire.NewSet(NewData)
 
+// Data .
 type Data struct {
 	db *gorm.DB
 }
 
+// NewData .
 func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
-	must.Same(c.Database.Driver, "sqlite3")
-	db := rese.P1(gorm.Open(sqlite.Open(c.Database.Source), &gorm.Config{}))
+	dsn := must.Nice(c.Database.Source)
+	db := rese.P1(gorm.Open(sqlite.Open(dsn), &gorm.Config{
+		Logger: loggergorm.Default.LogMode(loggergorm.Info),
+	}))
+
+	// Check if migration scripts are missing
+	// 检查是否缺少迁移脚本
+	checkmigration.CheckMigrate(db, models.Objects())
+
 	cleanup := func() {
 		log.NewHelper(logger).Info("closing the data resources")
-		_ = rese.P1(db.DB()).Close()
+		must.Done(rese.P1(db.DB()).Close())
 	}
-	return &Data{db: db}, cleanup, nil
+	return &Data{
+		db: db,
+	}, cleanup, nil
+}
+
+// DB returns the gorm database instance
+// 返回 gorm 数据库实例
+func (d *Data) DB() *gorm.DB {
+	return d.db
 }
```

## internal/pkg/appcfg/app_cfg.go (+29 -0)

```diff
@@ -0,0 +1,29 @@
+package appcfg
+
+import (
+	"github.com/go-kratos/kratos/v2/config"
+	"github.com/go-kratos/kratos/v2/config/file"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
+	"github.com/yylego/rese"
+)
+
+// ParseConfig parses config file and returns Bootstrap config
+// 解析配置文件并返回 Bootstrap 配置
+func ParseConfig(configPath string) *conf.Bootstrap {
+	c := config.New(
+		config.WithSource(
+			file.NewSource(configPath),
+		),
+	)
+	defer rese.F0(c.Close)
+
+	if err := c.Load(); err != nil {
+		panic(err)
+	}
+
+	var cfg conf.Bootstrap
+	if err := c.Scan(&cfg); err != nil {
+		panic(err)
+	}
+	return &cfg
+}
```

## internal/pkg/models/objects.go (+9 -0)

```diff
@@ -0,0 +1,9 @@
+package models
+
+// Objects returns all GORM model objects for migration
+// 返回所有用于迁移的 GORM 模型对象
+func Objects() []any {
+	return []any{
+		&Student{},
+	}
+}
```

## internal/pkg/models/student.go (+16 -0)

```diff
@@ -0,0 +1,16 @@
+package models
+
+import "gorm.io/gorm"
+
+// Student represents a student database model
+// 学生数据库模型
+type Student struct {
+	gorm.Model
+	Name string `gorm:"type:varchar(255)"`
+}
+
+// TableName returns the table name
+// 返回表名
+func (*Student) TableName() string {
+	return "students"
+}
```

## scripts/20260314101144_create_table.down.sql (+5 -0)

```diff
@@ -0,0 +1,5 @@
+-- reverse -- CREATE INDEX `idx_students_deleted_at` ON `students`(`deleted_at`);
+DROP INDEX IF EXISTS `idx_students_deleted_at`;
+
+-- reverse -- CREATE TABLE `students` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` varchar(255));
+DROP TABLE IF EXISTS `students`;
```

## scripts/20260314101144_create_table.up.sql (+10 -0)

```diff
@@ -0,0 +1,10 @@
+CREATE TABLE `students`
+(
+    `id`         integer PRIMARY KEY AUTOINCREMENT,
+    `created_at` datetime,
+    `updated_at` datetime,
+    `deleted_at` datetime,
+    `name`       varchar(255)
+);
+
+CREATE INDEX `idx_students_deleted_at` ON `students` (`deleted_at`);
```

