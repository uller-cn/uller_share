--
-- SQLiteStudio v3.4.4 生成的文件，周五 3月 28 17:28:37 2025
--
-- 所用的文本编码：UTF-8
--
PRAGMA foreign_keys = off;
BEGIN TRANSACTION;

-- 表：download_history
DROP TABLE IF EXISTS download_history;

CREATE TABLE IF NOT EXISTS download_history (
    history_id  INTEGER        UNIQUE
                               NOT NULL,
    share_id    INTEGER        NOT NULL,
    title       VARCHAR (200)  NOT NULL,
    local_path  VARCHAR (1000) NOT NULL,
    ip          VARCHAR (128)  NOT NULL,
    ext         VARCHAR (32),
    size        INTEGER        NOT NULL
                               DEFAULT (0),
    finish      INTEGER        NOT NULL
                               DEFAULT (0),
    status      INTEGER (2)    NOT NULL
                               DEFAULT (1),
    create_time DATETIME       NOT NULL
                               DEFAULT (CURRENT_TIMESTAMP),
    update_time DATETIME       NOT NULL
                               DEFAULT (CURRENT_TIMESTAMP) 
);


-- 表：profile
DROP TABLE IF EXISTS profile;

CREATE TABLE IF NOT EXISTS profile (
    nick              VARCHAR (32),
    last_share_folder VARCHAR (1000),
    qr_time           INTEGER        DEFAULT ( -1) 
                                     NOT NULL,
    download_routine  INTEGER (4)    NOT NULL
                                     DEFAULT (3),
    update_time       DATETIME       NOT NULL
                                     DEFAULT (CURRENT_TIMESTAMP) 
);

INSERT INTO profile values('','',0,3,CURRENT_TIMESTAMP);

-- 表：share
DROP TABLE IF EXISTS share;

CREATE TABLE IF NOT EXISTS share (
    share_id    INTEGER        PRIMARY KEY
                               UNIQUE
                               NOT NULL,
    title       VARCHAR (200),
    local_path  VARCHAR (1000),
    ext         VARCHAR (32),
    size        INTEGER (20)   DEFAULT (0),
    expire_time INTEGER (20)   NOT NULL
                               DEFAULT (0),
    create_time DATETIME       NOT NULL
                               DEFAULT (CURRENT_TIMESTAMP),
    update_time DATETIME       NOT NULL
                               DEFAULT (CURRENT_TIMESTAMP) 
);


-- 索引：download_history_status
DROP INDEX IF EXISTS download_history_status;

CREATE INDEX IF NOT EXISTS download_history_status ON download_history (
    status
);


-- 索引：download_history_title
DROP INDEX IF EXISTS download_history_title;

CREATE INDEX IF NOT EXISTS download_history_title ON download_history (
    title
);


-- 索引：share_ext
DROP INDEX IF EXISTS share_ext;

CREATE INDEX IF NOT EXISTS share_ext ON share (
    ext
);


-- 索引：share_title
DROP INDEX IF EXISTS share_title;

CREATE INDEX IF NOT EXISTS share_title ON share (
    title
);


COMMIT TRANSACTION;
PRAGMA foreign_keys = on;
