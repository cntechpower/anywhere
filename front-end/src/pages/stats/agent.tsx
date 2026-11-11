import { useMemo, useState } from "react";
import { Button, Input, Table, Tooltip } from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  ClusterOutlined,
  ReloadOutlined,
  SearchOutlined,
  WifiOutlined,
} from "@ant-design/icons";
import PageHeader from "../../components/PageHeader";
import StatCard from "../../components/StatCard";
import StatusBadge from "../../components/StatusBadge";
import EmptyState from "../../components/EmptyState";
import ErrorState from "../../components/ErrorState";
import LoadingState from "../../components/LoadingState";
import useApi from "../../hooks/useApi";
import useMediaQuery from "../../hooks/useMediaQuery";
import apis from "../../apis/apis";
import { apiGet } from "../../api/client";
import {
  formatDateTime,
  formatNumber,
  parseServerTime,
  timeAgo,
} from "../../utils/format";
import type { AgentInfo } from "../../types";

const { Search } = Input;
const HEALTHY_WINDOW_MS = 60 * 1000;

function zoneKey(record: AgentInfo): string {
  return `${record.userName ?? ""}/${record.zoneName ?? ""}`;
}

function isHealthy(record: AgentInfo, now: number): boolean {
  const time = parseServerTime(record.lastAckRcv);
  if (!time) {
    return false;
  }
  return now - time.getTime() < HEALTHY_WINDOW_MS;
}

export default function AgentList() {
  const [query, setQuery] = useState("");
  const isMobile = useMediaQuery("(max-width: 860px)");
  const { data, error, loading, reload } = useApi<AgentInfo[]>(
    (signal) => apiGet<AgentInfo[]>(apis.proxyAgentListApi, { signal }),
    { pollMs: 10000 }
  );

  const list = useMemo(() => (Array.isArray(data) ? data : []), [data]);
  const now = Date.now();

  const filtered = useMemo(() => {
    if (!query) {
      return list;
    }
    const keyword = query.toLowerCase();
    return list.filter((record) =>
      [record.agentId, record.userName, record.zoneName, record.agentAdminAddr]
        .filter(Boolean)
        .join(" ")
        .toLowerCase()
        .includes(keyword)
    );
  }, [list, query]);

  const zones = new Set(list.map(zoneKey));
  const healthy = list.filter((record) => isHealthy(record, now)).length;
  const stale = list.length - healthy;

  const columns: ColumnsType<AgentInfo> = [
    {
      title: "节点",
      key: "agent",
      render: (_, record) => (
        <div className="cell-stack">
          <span className="cell-primary">{record.agentId || "-"}</span>
          <StatusBadge tone={isHealthy(record, now) ? "ok" : "bad"}>
            {isHealthy(record, now) ? "在线" : "心跳超时"}
          </StatusBadge>
        </div>
      ),
    },
    {
      title: "归属",
      key: "owner",
      render: (_, record) => (
        <div className="cell-stack">
          <span className="cell-primary">{record.userName || "-"}</span>
          <span className="cell-muted">{record.zoneName || "-"}</span>
        </div>
      ),
    },
    {
      title: "节点地址",
      dataIndex: "agentAdminAddr",
      key: "agentAdminAddr",
      render: (value: string | undefined) => (
        <span className="mono-text">{value || "-"}</span>
      ),
    },
    {
      title: "最近心跳",
      dataIndex: "lastAckRcv",
      key: "lastAckRcv",
      render: (value: string | undefined) => (
        <Tooltip title={formatDateTime(value)}>
          <span className="cell-muted">{timeAgo(value, now)}</span>
        </Tooltip>
      ),
    },
  ];

  const renderMobileCards = () => (
    <div className="stack">
      {filtered.map((record) => (
        <article className="conn-card" key={record.agentId}>
          <div className="row-between">
            <span className="cell-primary">{record.agentId}</span>
            <StatusBadge tone={isHealthy(record, now) ? "ok" : "bad"}>
              {isHealthy(record, now) ? "在线" : "心跳超时"}
            </StatusBadge>
          </div>
          <div className="divider" />
          <dl className="detail-list">
            <div className="detail-row">
              <dt>归属</dt>
              <dd>
                {record.userName} / {record.zoneName}
              </dd>
            </div>
            <div className="detail-row">
              <dt>节点地址</dt>
              <dd>{record.agentAdminAddr || "-"}</dd>
            </div>
            <div className="detail-row">
              <dt>最近心跳</dt>
              <dd>{timeAgo(record.lastAckRcv, now)}</dd>
            </div>
          </dl>
        </article>
      ))}
    </div>
  );

  return (
    <>
      <PageHeader
        eyebrow="Agents"
        title="节点"
        description="Agent 通过 mTLS 长连接接入，每秒上报心跳。超过 60 秒未上报将被标记为心跳超时。"
        actions={
          <Button icon={<ReloadOutlined />} onClick={reload} loading={loading}>
            刷新
          </Button>
        }
      />

      {!data && loading && <LoadingState variant="cards" />}
      {!data && error && (
        <div className="panel">
          <ErrorState error={error} onRetry={reload} title="无法读取节点列表" />
        </div>
      )}

      {data && (
        <>
          <div className="grid stats-grid" style={{ marginBottom: 20 }}>
            <StatCard
              label="节点总数"
              value={formatNumber(list.length)}
              hint="已注册的控制连接"
              icon={<ClusterOutlined />}
            />
            <StatCard
              label="可用区"
              value={formatNumber(zones.size)}
              hint="用户 / 可用区组合"
              icon={<WifiOutlined />}
              tone="info"
            />
            <StatCard
              label="在线"
              value={formatNumber(healthy)}
              hint="60 秒内上报过心跳"
              icon={<WifiOutlined />}
            />
            <StatCard
              label="心跳超时"
              value={formatNumber(stale)}
              hint="需要检查 Agent 状态"
              icon={<WifiOutlined />}
              tone={stale > 0 ? "danger" : ""}
            />
          </div>

          <section className="panel">
            <div className="panel-header">
              <div>
                <h2 className="panel-title">节点列表</h2>
                <p className="panel-subtitle">
                  共 {formatNumber(list.length)} 个节点
                </p>
              </div>
              <Search
                allowClear
                className="search"
                placeholder="搜索节点、用户、可用区或地址"
                prefix={<SearchOutlined />}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                aria-label="搜索节点"
              />
            </div>
            <div className="panel-body flush">
              {filtered.length === 0 ? (
                <EmptyState
                  icon={<ClusterOutlined />}
                  title={
                    list.length === 0 ? "还没有 Agent 接入" : "没有匹配的节点"
                  }
                  description={
                    list.length === 0
                      ? "在目标机器上启动 Agent 并连接到服务端后，节点会自动出现在这里。"
                      : "尝试调整搜索关键词。"
                  }
                />
              ) : isMobile ? (
                <div className="panel-body">{renderMobileCards()}</div>
              ) : (
                <Table<AgentInfo>
                  rowKey={(record) => record.agentId ?? ""}
                  columns={columns}
                  dataSource={filtered}
                  pagination={false}
                  scroll={{ x: 680 }}
                />
              )}
            </div>
          </section>
        </>
      )}
    </>
  );
}
