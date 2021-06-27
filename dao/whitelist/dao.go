package whitelist

import (
	"context"
	"fmt"
	"time"

	"github.com/cntechpower/anywhere/dao"

	"github.com/cntechpower/utils/log"
	_ "github.com/go-sql-driver/mysql"
)

var header = log.NewHeader("whitelist")

type DenyItem struct {
	Ip      string
	Address string
	Count   int64
}

/*
create table if not exists whitelist_deny_history(
  id int AUTO_INCREMENT COMMENT '自增ID',
  ip varchar(15) NOT NULL COMMENT '被拒绝的IP地址',
  remote_port int NOT NULL DEFAULT 0 COMMENT '外网端口',
  agent_id varchar(50) NOT NULL COMMENT '节点名',
  local_addr varchar(25) NOT NULL COMMENT '内网地址',
  ctime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  mtime timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  address_en varchar(30) NOT NULL DEFAULT '',
  address_cn varchar(30) NOT NULL DEFAULT '',
  time_zone varchar(15) NOT NULL DEFAULT '',
  country_code varchar(5) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY ix_mtime (mtime),
  KEY idx_ip (ip)
) ENGINE = InnoDB DEFAULT CHARSET = utf8 COMMENT = '防火墙拦截记录表';`

*/
const (
	insertWhiteListHistorySql = `
insert into
  whitelist_deny_history(remote_port, ip, agent_id, local_addr)
values(?, ?, ?, ?)
`

	sqlTotalDenyRankDetail = `
SELECT ip, 
       address_cn, 
       count(*) AS deny_count 
FROM   whitelist_deny_history 
GROUP  BY ip, 
          address_cn 
ORDER  BY deny_count DESC limit ?; 
`

	sqlTotalDenyRankDetailCount = `
SELECT count(*) AS deny_count 
FROM   whitelist_deny_history; 
`

	sqlTotalDenyRankIpCount = `
SELECT count(DISTINCT( ip )) 
FROM   whitelist_deny_history; 
`

	sqlDailyDenyRankDetail = `
SELECT ip, 
       address_cn, 
       count(*) AS deny_count 
FROM   whitelist_deny_history 
WHERE  date(ctime) = curdate() 
GROUP  BY ip, 
          address_cn 
ORDER  BY deny_count DESC limit ?; 
`

	sqlDailyDenyRankDetailCount = `
SELECT count(*) AS deny_count 
FROM   whitelist_deny_history
WHERE  date(ctime) = curdate();
`

	sqlDailyDenyRankIpCount = `
SELECT count(DISTINCT( ip )) 
FROM   whitelist_deny_history
WHERE  date(ctime) = curdate(); 
`
)

const (
	RankTypeTotal = "total"
	RankTypeDaily = "daily"
)

type denyRankSqls struct {
	detailSql      string
	ipCountSql     string
	detailCountSql string
}

var denyRankSqlMap = map[string]*denyRankSqls{
	RankTypeTotal: {
		detailSql:      sqlTotalDenyRankDetail,
		detailCountSql: sqlTotalDenyRankDetailCount,
		ipCountSql:     sqlTotalDenyRankIpCount,
	},
	RankTypeDaily: {
		detailSql:      sqlDailyDenyRankDetail,
		detailCountSql: sqlDailyDenyRankDetailCount,
		ipCountSql:     sqlDailyDenyRankIpCount,
	},
}

func AddWhiteListDenyIp(remotePort int, agentId, localAddr, ip string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	_, err := dao.MySQL().ExecContext(ctx, insertWhiteListHistorySql, remotePort, ip, agentId, localAddr)
	cancel()
	if err != nil {
		header.Errorf("save whitelist history error: %v", err)
	}
	return err
}

func GetWhiteListDenyRank(typ string, limit int64) (details []*DenyItem, detailCount, ipCount int64, err error) {
	sqls := denyRankSqlMap[typ]
	if sqls == nil {
		err = fmt.Errorf("no such deny rank type")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err = dao.MySQL().QueryRow(sqls.detailCountSql).Scan(&detailCount)
	if err != nil {
		return
	}
	err = dao.MySQL().QueryRow(sqls.ipCountSql).Scan(&ipCount)
	if err != nil {
		return
	}
	rows, err := dao.MySQL().QueryContext(ctx, sqls.detailSql, limit)
	if err != nil {
		header.Errorf("query total deny rank error: %v", err)
		return
	}
	details = make([]*DenyItem, 0)
	for rows.Next() {
		i := &DenyItem{}
		if err = rows.Scan(&i.Ip, &i.Address, &i.Count); err != nil {
			return
		}
		details = append(details, i)
	}
	return
}
