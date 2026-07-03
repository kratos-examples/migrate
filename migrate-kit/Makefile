.PHONY: all inc dec status

# run each pending migration
# 向前跑完所有待执行的迁移
all:
	go run . migrate all

# step one migration forward
# 向前迁移一步
inc:
	go run . migrate inc

# step one migration back
# 向后回退一步
dec:
	go run . migrate dec

# show the current schema version
# 显示当前表结构版本
status:
	go run . migrate
