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
@@ -1,20 +1,23 @@
 package main
 
 import (
-	"flag"
+	"fmt"
 	"log/slog"
 	"os"
 
 	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
 	"github.com/go-kratos/kratos/v3"
-	"github.com/go-kratos/kratos/v3/config"
-	"github.com/go-kratos/kratos/v3/config/file"
 	"github.com/go-kratos/kratos/v3/log"
 	"github.com/go-kratos/kratos/v3/transport/grpc"
 	"github.com/go-kratos/kratos/v3/transport/http"
+	"github.com/spf13/cobra"
 	"github.com/yylego/done"
+	"github.com/yylego/kratos-examples/demo1kratos/cmd/demo1kratos/cfgpath"
+	"github.com/yylego/kratos-examples/demo1kratos/cmd/demo1kratos/subcmds"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/appcfg"
 	"github.com/yylego/must"
+	"github.com/yylego/must/mustslice"
 	"github.com/yylego/rese"
 
 	_ "go.uber.org/automaxprocs"
@@ -26,12 +29,10 @@
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
 
 func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
@@ -49,7 +50,6 @@
 }
 
 func main() {
-	flag.Parse()
 	logger := log.NewLogger(
 		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
 			AddSource: true,
@@ -62,18 +62,35 @@
 		slog.String("service.version", Version),
 	)
 	log.SetDefault(logger)
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
+func runApp(cfg *conf.Bootstrap, logger *slog.Logger) {
 	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
 	defer cleanup()
 
```

## cmd/demo1kratos/subcmds/sub_cmds.go (+118 -0)

```diff
@@ -0,0 +1,118 @@
+package subcmds
+
+import (
+	"log/slog"
+
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
+func NewVersionCmd(serviceName, version string, logger *slog.Logger) *cobra.Command {
+	return &cobra.Command{
+		Use:   "version",
+		Short: "Print version info",
+		Run: func(cmd *cobra.Command, args []string) {
+			logger.Info("version info", "service-name", serviceName, "version", version)
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
+func NewMigrateCmd(logger *slog.Logger) *cobra.Command {
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

## cmd/demo1kratos/wire_gen.go (+1 -5)

```diff
@@ -28,11 +28,7 @@
 	if err != nil {
 		return nil, nil, err
 	}
-	studentUsecase, err := biz.NewStudentUsecase(dataData, logger)
-	if err != nil {
-		cleanup()
-		return nil, nil, err
-	}
+	studentUsecase := biz.NewStudentUsecase(dataData, logger)
 	studentService := service.NewStudentService(studentUsecase)
 	grpcServer := server.NewGRPCServer(confServer, studentService, logger)
 	httpServer := server.NewHTTPServer(confServer, studentService, logger)
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
-    driver: postgres
-    source: host=localhost port=5432 user=postgres password=123 dbname=kratos_examples_db sslmode=disable
+    driver: sqlite
+    source: ./bin/demo1kratos.db
```

## internal/biz/student.go (+48 -108)

```diff
@@ -2,152 +2,92 @@
 
 import (
 	"context"
-	"errors"
 	"log/slog"
 
+	"github.com/brianvoe/gofakeit/v7"
+	"github.com/go-kratos/kratos/v3/errors"
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
-	"github.com/yylego/must"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
 	"gorm.io/gorm"
-	"gorm.io/gorm/clause"
 )
 
-// Student is the GORM type mapped to the "students" table.
-//
-// Student 是映射到 students 表的 GORM 模型
 type Student struct {
-	ID        int64  `gorm:"primaryKey;autoIncrement"`
-	Name      string `gorm:"size:128;not null"`
+	ID        int64
+	Name      string
 	Age       int32
-	ClassName string `gorm:"size:128"`
+	ClassName string
 }
 
-func (Student) TableName() string { return "students" }
-
-// The mirrored Article type behind cascade-delete lives in article.go.
-// 用于级联删除的 Article 镜像模型定义在 article.go 中。
-
 type StudentUsecase struct {
 	data *data.Data
-	slog *slog.Logger
+	log  *slog.Logger
 }
 
-func NewStudentUsecase(data *data.Data, logger *slog.Logger) (*StudentUsecase, error) {
-	// Share one database with the article service: keep both tables in sync here
-	// 与文章服务共用一个库：在这里把两张表都建好
-	if err := data.DB().AutoMigrate(&Student{}, &Article{}); err != nil {
-		return nil, err
-	}
-	return &StudentUsecase{data: data, slog: logger}, nil
+func NewStudentUsecase(data *data.Data, logger *slog.Logger) *StudentUsecase {
+	return &StudentUsecase{data: data, log: logger}
 }
 
 func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
-	must.Nice(s.Name)
+	db := uc.data.DB()
 
-	res := &Student{Name: s.Name, Age: s.Age, ClassName: s.ClassName}
-	if err := uc.data.DB().WithContext(ctx).Create(res).Error; err != nil {
-		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("create student: %v", err))
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
 	}
-	uc.slog.InfoContext(ctx, "created student", "id", res.ID, "name", res.Name)
-	return res, nil
+
+	var res Student
+	if err := gofakeit.Struct(&res); err != nil {
+		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("fake: %v", err))
+	}
+	res.ID = s.ID
+	res.Name = s.Name
+	return &res, nil
 }
 
 func (uc *StudentUsecase) UpdateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
-	must.True(s.ID > 0)
-	must.Nice(s.Name)
-
-	res := &Student{ID: s.ID}
-	upd := uc.data.DB().WithContext(ctx).Model(res).Updates(map[string]any{
-		"name":       s.Name,
-		"age":        s.Age,
-		"class_name": s.ClassName,
-	})
-	if upd.Error != nil {
-		return nil, ebzkratos.New(pb.ErrorDbError("update student: %v", upd.Error))
+	var res Student
+	if err := gofakeit.Struct(&res); err != nil {
+		return nil, ebzkratos.New(pb.ErrorServerError("fake: %v", err))
 	}
-	if upd.RowsAffected == 0 {
-		return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", s.ID))
-	}
-	if err := uc.data.DB().WithContext(ctx).First(res, s.ID).Error; err != nil {
-		return nil, ebzkratos.New(pb.ErrorDbError("reload student: %v", err))
-	}
-	return res, nil
+	return &res, nil
 }
 
 func (uc *StudentUsecase) DeleteStudent(ctx context.Context, id int64) *ebzkratos.Ebz {
-	must.True(id > 0)
-
-	// Atomic, race-safe cascade delete, in one transaction:
-	//   1. lock the student row (FOR UPDATE) so no article can target
-	//      this student meanwhile — CreateArticle takes a conflicting FOR SHARE
-	//      lock on the same row, so the two operations serialize;
-	//   2. delete the student's articles (children first);
-	//   3. delete the student (parent last).
-	// 原子且并发安全的级联删除，全部在一个事务里完成：
-	//   ① 用 FOR UPDATE 锁住学生行，删除期间不允许给该学生并发新建文章——CreateArticle
-	//      会对同一行加互斥的 FOR SHARE 锁，二者因此串行化；
-	//   ② 先删该学生名下的文章（子表在前）；
-	//   ③ 再删学生本身（父表在后）。
-	var notFound bool
-	var removedArticles int64
-	err := uc.data.DB().WithContext(ctx).Transaction(func(db *gorm.DB) error {
-		var s Student
-		if err := db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).First(&s, id).Error; err != nil {
-			if errors.Is(err, gorm.ErrRecordNotFound) {
-				notFound = true
-				return nil
-			}
-			return err
-		}
-		del := db.Where("student_id = ?", id).Delete(&Article{})
-		if del.Error != nil {
-			return del.Error
-		}
-		removedArticles = del.RowsAffected
-		return db.Delete(&Student{}, id).Error
-	})
-	if err != nil {
-		return ebzkratos.New(pb.ErrorTxError("delete student with articles: %v", err))
-	}
-	if notFound {
-		return ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
-	}
-	uc.slog.InfoContext(ctx, "deleted student and cascaded articles", "student_id", id, "articles_removed", removedArticles)
 	return nil
 }
 
 func (uc *StudentUsecase) GetStudent(ctx context.Context, id int64) (*Student, *ebzkratos.Ebz) {
-	must.True(id > 0)
+	db := uc.data.DB()
 
-	res := &Student{}
-	if err := uc.data.DB().WithContext(ctx).First(res, id).Error; err != nil {
+	var student models.Student
+	if err := db.WithContext(ctx).First(&student, id).Error; err != nil {
 		if errors.Is(err, gorm.ErrRecordNotFound) {
-			return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
+			return nil, ebzkratos.New(pb.ErrorServerError("not found: %v", err))
 		}
-		return nil, ebzkratos.New(pb.ErrorDbError("get student: %v", err))
+		return nil, ebzkratos.New(pb.ErrorServerError("db: %v", err))
 	}
-	return res, nil
+
+	return &Student{
+		ID:   int64(student.ID),
+		Name: student.Name,
+	}, nil
 }
 
 func (uc *StudentUsecase) ListStudents(ctx context.Context, page int32, pageSize int32) ([]*Student, int32, *ebzkratos.Ebz) {
-	if page < 1 {
-		page = 1
-	}
-	if pageSize < 1 {
-		pageSize = 10
-	}
-
-	db := uc.data.DB().WithContext(ctx)
-
-	var total int64
-	if err := db.Model(&Student{}).Count(&total).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("count students: %v", err))
-	}
-
 	var items []*Student
-	if err := db.Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("list students: %v", err))
-	}
-	return items, int32(total), nil
+	gofakeit.Slice(&items)
+	return items, int32(len(items)), nil
 }
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

## internal/data/data.go (+26 -12)

```diff
@@ -4,32 +4,46 @@
 	"log/slog"
 
 	"github.com/google/wire"
+	"github.com/yylego/go-migrate/checkmigration"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
-	"gorm.io/driver/postgres"
+	"gorm.io/driver/sqlite"
 	"gorm.io/gorm"
+	loggergorm "gorm.io/gorm/logger"
 )
 
+// ProviderSet is data providers.
 var ProviderSet = wire.NewSet(NewData)
 
+// Data .
 type Data struct {
 	db *gorm.DB
 }
 
-// DB exposes the underlying gorm handle so the biz code can run true queries.
-//
-// DB 暴露底层 gorm 句柄，供 biz 层执行真实的数据库读写
-func (d *Data) DB() *gorm.DB {
-	return d.db
-}
-
+// NewData .
 func NewData(c *conf.Data, logger *slog.Logger) (*Data, func(), error) {
-	must.Same(c.Database.Driver, "postgres")
-	db := rese.P1(gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{}))
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
 		logger.Info("closing the data resources")
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
+	"github.com/go-kratos/kratos/v3/config"
+	"github.com/go-kratos/kratos/v3/config/file"
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

