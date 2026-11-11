import { Button, Tooltip } from "antd";
import {
  ApiOutlined,
  ClusterOutlined,
  LinkOutlined,
  ReloadOutlined,
  StopOutlined,
  SwapOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import { Link } from "react-router-dom";
import PageHeader from "../components/PageHeader";
import StatCard from "../components/StatCard";
import StatusBadge from "../components/StatusBadge";
import BarList from "../components/BarList";
import ErrorState from "../components/ErrorState";
import LoadingState from "../components/LoadingState";
import useApi from "../hooks/useApi";
import apis from "../apis/apis";
import { apiGet } from "../api/client";
import { formatBytes, formatDateTime, formatNumber } from "../utils/format";
import type { BarItem, ProxyConfig, Summary } from "../types";

function topKey(item: ProxyConfig, index: number): string {
  return `${item.user_name}/${item.zone_name}/${item.remote_port}/${index}`;
}

export default function Home() {
  const { data, error, loading, reload, updatedAt } = useApi<Summary>(
    (signal) => apiGet<Summary>(apis.getSummaryApi, { signal }),
    { pollMs: 10000 }
  );

  const summary = data ?? {};
  const rejectTop: BarItem[] = (
    summary.proxy_connect_reject_count_top10 ?? []
  ).map((item, index) => ({
    key: topKey(item, index),
    label: `${item.user_name ?? "-"} / ${item.zone_name ?? "-"} · :${item.remote_port ?? "-"}`,
    value: Number(item.proxy_connect_reject_count) || 0,
  }));

  const flowTop: BarItem[] = (summary.proxy_network_flow_top10 ?? []).map(
    (item, index) => ({
      key: topKey(item, index),
      label: `${item.local_addr ?? "-"} → :${item.remote_port ?? "-"}`,
      value:
        (Number(item.network_flow_local_to_remote_in_bytes) || 0) +
        (Number(item.network_flow_remote_to_local_in_bytes) || 0),
    })
  );

  const hasData = Boolean(data);

  return (
    <>
      <PageHeader
        eyebrow="Operations overview"
        title="概览"
        description="所有可用区的运行状态、累计流量与拒绝连接排名，每 10 秒自动刷新。"
        actions={
          <>
            <StatusBadge tone={error ? "warn" : "ok"}>
              {error ? "数据可能已过期" : "服务运行中"}
            </StatusBadge>
            <Button icon={<ReloadOutlined />} onClick={reload} loading={loading}>
              刷新
            </Button>
            <Link to="/configs/add">
              <Button type="primary" icon={<ThunderboltOutlined />}>
                新建配置
              </Button>
            </Link>
          </>
        }
      />

      {!hasData && loading && <LoadingState variant="cards" />}
      {!hasData && error && (
        <div className="panel">
          <ErrorState error={error} onRetry={reload} title="无法读取运行概览" />
        </div>
      )}

      {hasData && (
        <>
          <div className="grid stats-grid">
            <StatCard
              label="Agent 节点"
              value={formatNumber(summary.agent_total_count)}
              hint="已注册的控制连接"
              icon={<ClusterOutlined />}
            />
            <StatCard
              label="代理配置"
              value={formatNumber(summary.proxy_config_total_count)}
              hint="TCP / UDP 监听端口"
              icon={<ApiOutlined />}
              tone="info"
            />
            <StatCard
              label="活跃连接"
              value={formatNumber(summary.current_proxy_connection_count)}
              hint="当前正在转发的会话"
              icon={<LinkOutlined />}
            />
            <StatCard
              label="拒绝连接"
              value={formatNumber(summary.proxy_connect_reject_count)}
              hint="白名单或限流拦截"
              icon={<StopOutlined />}
              tone={
                Number(summary.proxy_connect_reject_count) > 0 ? "danger" : ""
              }
            />
            <StatCard
              label="累计流量"
              value={formatBytes(summary.network_flow_total_count_in_bytes)}
              hint="进 + 出双向累计"
              icon={<SwapOutlined />}
              tone="warning"
            />
          </div>

          <div className="grid two-col" style={{ marginBottom: 20 }}>
            <section className="panel">
              <div className="panel-header">
                <div>
                  <h2 className="panel-title">流量 Top 10</h2>
                  <p className="panel-subtitle">按进 + 出累计流量排序</p>
                </div>
                <Tooltip title="数据来源：服务端每 5 秒聚合一次">
                  <span className="tag-mono">top-10</span>
                </Tooltip>
              </div>
              <div className="panel-body">
                <BarList
                  items={flowTop}
                  valueFormatter={formatBytes}
                  emptyText="暂无流量数据"
                />
              </div>
            </section>

            <section className="panel">
              <div className="panel-header">
                <div>
                  <h2 className="panel-title">拒绝连接 Top 10</h2>
                  <p className="panel-subtitle">白名单拦截与并发上限拒绝</p>
                </div>
                <span className="tag-mono">top-10</span>
              </div>
              <div className="panel-body">
                <BarList
                  items={rejectTop}
                  tone="danger"
                  valueFormatter={(value) => `${formatNumber(value)} 次`}
                  emptyText="暂无拒绝记录"
                  emptyDescription="没有客户端被白名单或并发限制拒绝。"
                />
              </div>
            </section>
          </div>

          <div className="grid two-col">
            <section className="panel">
              <div className="panel-header">
                <div>
                  <h2 className="panel-title">运行信息</h2>
                  <p className="panel-subtitle">最近一次聚合结果</p>
                </div>
              </div>
              <div className="panel-body">
                <dl className="detail-list">
                  <div className="detail-row">
                    <dt>数据刷新时间</dt>
                    <dd>{formatDateTime(updatedAt)}</dd>
                  </div>
                  <div className="detail-row">
                    <dt>累计连接数</dt>
                    <dd>{formatNumber(summary.proxy_connect_total_count)}</dd>
                  </div>
                  <div className="detail-row">
                    <dt>累计拒绝数</dt>
                    <dd>{formatNumber(summary.proxy_connect_reject_count)}</dd>
                  </div>
                  <div className="detail-row">
                    <dt>自动刷新</dt>
                    <dd>每 10 秒</dd>
                  </div>
                </dl>
              </div>
            </section>

            <section className="panel">
              <div className="panel-header">
                <div>
                  <h2 className="panel-title">快捷操作</h2>
                  <p className="panel-subtitle">常用入口</p>
                </div>
              </div>
              <div className="panel-body">
                <div className="quick-actions">
                  <Link to="/configs/add">
                    <Button type="primary">新增代理配置</Button>
                  </Link>
                  <Link to="/configs/list">
                    <Button>配置列表</Button>
                  </Link>
                  <Link to="/stats/agents">
                    <Button>查看节点</Button>
                  </Link>
                  <Link to="/stats/conns">
                    <Button>活跃连接</Button>
                  </Link>
                </div>
                <div className="divider" />
                <p className="cell-muted" style={{ margin: 0 }}>
                  配置变更会立即下发给可用区内的 Agent；连接数、流量等指标按 10
                  秒周期刷新。
                </p>
              </div>
            </section>
          </div>
        </>
      )}
    </>
  );
}
