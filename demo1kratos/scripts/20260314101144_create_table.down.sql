-- reverse -- CREATE INDEX `idx_students_deleted_at` ON `students`(`deleted_at`);
DROP INDEX IF EXISTS `idx_students_deleted_at`;

-- reverse -- CREATE TABLE `students` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`name` varchar(255));
DROP TABLE IF EXISTS `students`;
