CREATE TABLE students
(
    id         bigserial PRIMARY KEY,
    name       varchar(128) NOT NULL,
    age        integer,
    class_name varchar(128)
);

CREATE TABLE articles
(
    id         bigserial PRIMARY KEY,
    title      varchar(256) NOT NULL,
    content    text,
    student_id bigint
);

CREATE INDEX idx_articles_student_id ON articles (student_id);
