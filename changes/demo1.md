# Changes

Code differences compared to source project.

## internal/biz/student.go (+13 -5)

```diff
@@ -5,6 +5,7 @@
 	"errors"
 	"log/slog"
 
+	"github.com/yylego/go-migrate/checkmigration"
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
@@ -34,11 +35,18 @@
 }
 
 func NewStudentUsecase(data *data.Data, logger *slog.Logger) (*StudentUsecase, error) {
-	// Share one database with the article service: keep both tables in sync here
-	// 与文章服务共用一个库：在这里把两张表都建好
-	if err := data.DB().AutoMigrate(&Student{}, &Article{}); err != nil {
-		return nil, err
-	}
+	// AutoMigrate is disabled on purpose: the schema comes from the sibling migrate-kit
+	// project via hand-written go-migrate scripts. Here we just check that the live
+	// schema matches the models and crash fast when a migration is pending, since the
+	// sibling runs the migration.
+	//
+	// 有意停用 AutoMigrate：表结构改由旁挂的 migrate-kit 子项目用 go-migrate 脚本化迁移管理。
+	// 这里只检查实时表结构是否还与模型匹配，有待迁移则直接断言崩溃（真正的迁移请运行 migrate-kit）。
+	//
+	// if err := data.DB().AutoMigrate(&Student{}, &Article{}); err != nil {
+	// 	return nil, err
+	// }
+	must.Length(checkmigration.CheckMigrate(data.DB(), []any{&Student{}, &Article{}}), 0)
 	return &StudentUsecase{data: data, slog: logger}, nil
 }
 
```

