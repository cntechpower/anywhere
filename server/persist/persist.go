package persist

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cntechpower/anywhere/log"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB
var header *log.Header

type WhiteListDenyItem struct {
	Ip    string
	Count int64
}

const (
	createTableSql = `
create table if not exists whitelist_deny_history(
  id int AUTO_INCREMENT COMMENT '自增ID',
  ip varchar(15) NOT NULL COMMENT '被拒绝的IP地址',
  remote_port int NOT NULL DEFAULT 0 COMMENT '外网端口',
  agent_id varchar(50) NOT NULL COMMENT '节点名',
  local_addr varchar(25) NOT NULL COMMENT '内网地址',
  ctime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  mtime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (id),
  KEY ix_mtime (mtime)
) ENGINE = InnoDB DEFAULT CHARSET = utf8 COMMENT = '防火墙拦截记录表';`

	insertWhiteListHistorySql = `
insert into
  whitelist_deny_history(remote_port, ip, agent_id, local_addr)
values(?, ?, ?, ?)
`

	totalDenyRankSql = `
select
ip,
count(*) as deny_count
from
whitelist_deny_history
group by
ip
order by
deny_count desc;
`
)

func Init(dsn string) {
	if dsn == "" {
		panic(fmt.Errorf("mysql dsn is empty"))
	}
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	DB.SetConnMaxLifetime(time.Minute * 3)
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(10)
	_, err = DB.Exec(createTableSql)
	if err != nil {
		panic(err)
	}
	header = log.NewHeader("persist")
}

func AddWhiteListDenyIp(remotePort int, agentId, localAddr, ip string) error {
	if DB == nil {
		log.Errorf(header, "db is not init")
		return fmt.Errorf("db is not init")
	}
	_, err := DB.Exec(insertWhiteListHistorySql, remotePort, ip, agentId, localAddr)
	if err != nil {
		log.Errorf(header, "save whitelist history error: %v", err)
	}
	return err
}

func GetTotalDenyRank() (res []*WhiteListDenyItem, err error) {
	if DB == nil {
		log.Errorf(header, "db is not init")
		return nil, fmt.Errorf("db is not init")
	}
	rows, err := DB.Query(totalDenyRankSql)
	if err != nil {
		return nil, err
	}
	res = make([]*WhiteListDenyItem, 0)
	for rows.Next() {
		i := &WhiteListDenyItem{}
		if err := rows.Scan(&i.Ip, &i.Count); err != nil {
			return nil, err
		}
		res = append(res, i)
	}
	return res, nil
}
