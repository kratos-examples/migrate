CREATE TABLE `students`
(
    `id`         integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `name`       varchar(255)
);

CREATE INDEX `idx_students_deleted_at` ON `students` (`deleted_at`);
