package sqlite

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"github.com/shockerli/cvt"
	"github.com/sirupsen/logrus"
	"github.com/uller_share/common"
	"net/url"
	"strings"
	"sync"
	"time"
)

var dbname = "uller_share.dll"
var Db DataBase
var IsRun bool

type DataBase struct {
	RunPath string
	mutex   sync.RWMutex
	*sqlx.DB
}

func Exec(sqlStr string) (err error) {
	Db.mutex.Lock()
	//prepare, err := Db.Db.Prepare(sqlStr)
	//_, err = prepare.Exec()
	_, err = Db.Exec("PRAGMA synchronous=OFF;")
	if err != nil {
		logrus.Error("sqlite.Exec：", err)
	}
	_, err = Db.Exec(sqlStr)
	Db.mutex.Unlock()
	if err != nil {
		logrus.Error("sqlite.Exec：", err)
	}
	return
}

func QueryProfile() (retData common.Profile) {
	sqlStr := "select * from profile limit 1"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logrus.Error("查询个人信息错误：", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&retData.Nick, &retData.LastShareFolder, &retData.QrTime, &retData.DownloadRoutine, &retData.UpdateTime)
		if err != nil {
			logrus.Error("查询个人信息绑定数据错误：", err)
			return
		}
	}
	return
}

func QueryShareList(title string) (retData []common.Share) {
	where := ""
	if title != "" {
		where = " where title like '%" + title + "%'"
	}
	sqlStr := "select share_id,title,local_path,'" + common.LocalIp.String() + "' as ip,ext,size,expire_time,create_time,update_time from share" + where
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logrus.Error("查询分享文件错误：", err)
		return
	}
	defer rows.Close()
	share := common.Share{}
	for rows.Next() {
		err = rows.Scan(&share.ShareId, &share.Title, &share.LocalPath, &share.Ip, &share.Ext, &share.Size, &share.ExpireTime, &share.CreateTime, &share.UpdateTime)
		if err != nil {
			logrus.Error("查询分享文件绑定数据错误：", err)
			return
		}
		retData = append(retData, share)
	}
	return
}

func QueryShare(id string) (retData common.Share) {
	if !common.IsNum(id) {
		return retData
	}
	where := " where id=" + id + ""
	sqlStr := "select share_id,title,ext from share" + where
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logrus.Error("查询单个共享文件错误：", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&retData.ShareId, &retData.Title, &retData.Ext)
		if err != nil {
			logrus.Error("查询单个共享文件绑定数据错误：", err)
			return
		}
	}
	return
}

func QueryDownloadHistoryList(title string, ext []string, sType uint8, pageSize int, pageNumber int) (retData common.DownLoadHistoryList) {
	where := " where 0=0 "
	if title != "" {
		where += " and title like '%" + title + "%'"
	}
	if len(ext) > 0 {
		where += " and ext in ('" + strings.Join(ext, ",") + "')"
	}
	if sType == 0 {
		where += " and share_id=0"
	} else {
		where += " and share_id>0"
	}
	sqlStr := "select history_id,share_id,title,local_path,ip,size,finish,ext,status,create_time,update_time from download_history " + where + " order by update_time desc limit " + cvt.String(pageSize*(pageNumber)) + "," + cvt.String(pageSize*(pageNumber+1)) + ""
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logrus.Error("查询下载历史错误：", err)
		return
	}
	defer rows.Close()
	downLoadHistory := common.DownLoadHistory{}
	for rows.Next() {
		err = rows.Scan(&downLoadHistory.HistoryId, &downLoadHistory.Share.ShareId, &downLoadHistory.Title, &downLoadHistory.LocalPath, &downLoadHistory.Ip, &downLoadHistory.Size, &downLoadHistory.Finish, &downLoadHistory.Ext, &downLoadHistory.Status, &downLoadHistory.CreateTime, &downLoadHistory.UpdateTime)
		if err != nil {
			logrus.Error("查询下载历史绑定数据错误：", err)
			return
		}
		retData.DownLoadHistory = append(retData.DownLoadHistory, downLoadHistory)
	}
	sqlCountStr := "select count(0) from download_history " + where + ""
	rows, err = Db.Query(sqlCountStr)
	if err != nil {
		logrus.Error("查询下载历史数据总量错误：", err)
		return
	}
	for rows.Next() {
		err = rows.Scan(&retData.Total)
		if err != nil {
			logrus.Error("查询下载历史数据总量绑定数据错误：", err)
			return
		}
	}
	return
}

func QueryDownloadHistoryIds(shareIds string) (retData []common.DownLoadHistory) {
	sqlStr := "select history_id,share_id,title,local_path,ip,ext,size,finish,status,create_time,update_time from download_history where history_id in (select history_id from (select max(update_time),history_id from download_history where share_id in (" + shareIds + ") group by share_id) a)"
	rows, err := Db.Query(sqlStr)
	if err != nil {
		logrus.Error("查询下载历史错误：", err)
		return
	}
	defer rows.Close()
	downLoadHistory := common.DownLoadHistory{}
	for rows.Next() {
		err = rows.Scan(&downLoadHistory.HistoryId, &downLoadHistory.Share.ShareId, &downLoadHistory.Title, &downLoadHistory.LocalPath, &downLoadHistory.Ip, &downLoadHistory.Ext, &downLoadHistory.Size, &downLoadHistory.Finish, &downLoadHistory.Status, &downLoadHistory.CreateTime, &downLoadHistory.UpdateTime)
		if err != nil {
			logrus.Error("查询下载历史绑定数据错误：", err)
			return
		}
		retData = append(retData, downLoadHistory)
	}
	return
}

func UpdateDownloadHistoryNoFinish(historyId string) (err error) {
	return Exec("update download_history set `status`=0,finish=0 where history_id=" + historyId + "")
}

func Open() {
	if Db.RunPath == "" {
		IsRun = false
		return
	}
	var err error
	key := url.QueryEscape(common.SqlitePwd)
	dbPath := Db.RunPath + "/" + dbname
	dbExists := common.FileExists(dbPath)
	dsn := fmt.Sprintf("%s?_pragma_key=%s&synchronous=OFF", dbPath, key)
	if !dbExists {
		err = CreatDB(dsn)
		if err != nil {
			logrus.Error("创建数据库错误：", err)
			IsRun = false
			return
		}
	}
	Db.DB, err = sqlx.Connect("sqlite3", dsn)
	if err != nil {
		logrus.Error("打开数据库错误：", err)
		IsRun = false
		return
	}
	// 设置连接池
	Db.SetConnMaxLifetime(4 * time.Hour)
	Db.SetMaxOpenConns(10)
	Db.SetMaxIdleConns(5)
	IsRun = true
	return
}

// 创建数据库
func CreatDB(dsn string) (err error) {
	Db.DB, err = sqlx.Open("sqlite3", dsn)
	defer Db.DB.Close()

	_, err = Db.Exec(`--
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
`)
	return
}
