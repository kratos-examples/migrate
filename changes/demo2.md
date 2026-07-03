# Changes

Code differences compared to source project.

## internal/biz/article.go (+13 -6)

```diff
@@ -5,6 +5,7 @@
 	"errors"
 	"log/slog"
 
+	"github.com/yylego/go-migrate/checkmigration"
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
@@ -34,12 +35,18 @@
 }
 
 func NewArticleUsecase(data *data.Data, logger *slog.Logger) (*ArticleUsecase, error) {
-	// Migrate the owned table plus the mirrored students table (needed in the
-	// existence check); both services share one database
-	// 建好本服务拥有的 articles 表，外加镜像的 students 表（供存在性校验用）
-	if err := data.DB().AutoMigrate(&Article{}, &Student{}); err != nil {
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
+	// if err := data.DB().AutoMigrate(&Article{}, &Student{}); err != nil {
+	// 	return nil, err
+	// }
+	must.Length(checkmigration.CheckMigrate(data.DB(), []any{&Article{}, &Student{}}), 0)
 	return &ArticleUsecase{data: data, slog: logger}, nil
 }
 
```

