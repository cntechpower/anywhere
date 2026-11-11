import { useMemo, useState } from "react";
import { Button, Input, Popconfirm, Table, message } from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  DisconnectOutlined,
  LinkOutlined,
  ReloadOutlined,
  SearchOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import PageHeader from "../../components/PageHeader";
import StatCard from "../../components/StatCard";
import EmptyState from "../../components/EmptyState";
import ErrorState from "../../components/ErrorState";
import LoadingState from "../../components/LoadingState";
import useApi from "../../hooks/useApi";
import useMediaQuery from "../../hooks/useMediaQuery";
import apis from "../../apis/apis";
import { apiGet, apiPostForm } from "../../api/client";
import { apiErrorMessage, formatNumber } from "../../utils/format";
import type { ConnItem, ProxyConfig } from "../../types";

const { Search } = Input;

function includes(record: ConnItem, keyword: string): boolean {
  if (!keyword) {
    return true;
  }
  return [
    record.id,
    record.src_remote_addr,
    record.src_local_addr,
    record.src_name,
    record.dst_name,
    record.dst_remote_addr,
    record.dst_local_addr,
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase()
    .includes(keyword.toLowerCase());
}

export default function ConnectionList() {
  const [query, setQuery] = useState("");
  const isMobile = useMediaQuery("(max-width: 860px)");
  const { data, error, loading, reload } = useApi<ConnItem[]>(
    (signal) => apiGet<ConnItem[]>(apis.statsGetConnectionListApi, { signal }),
    { pollMs: 5000 }
  );
  const configsApi = useApi<ProxyConfig[]>(
    (signal) => apiGet<ProxyConfig[]>(apis.proxyConfigListApi, { signal }),
    { pollMs: 30000 }
  );

  // Keep the array identity stable so downstream useMemo hooks are not
  // invalidated on every render.
  const list = useMemo(() => (Array.isArray(data) ? data : []), [data]);
  const configs = useMemo(
    () => (Array.isArray(configsApi.data) ? configsApi.data : []),
    [configsApi.data]
  );
  const targetByPort = useMemo(() => {
    const map: Record<string, string> = {};
    configs.forEach((config) => {
      map[String(config.remote_port)] = config.local_addr ?? "-";
    });
    return map;
  }, [configs]);

  const targetOf = (record: ConnItem): string => {
    const addr = String(record.src_local_addr ?? "");
    const port = addr.slice(addr.lastIndexOf(":") + 1);
    return targetByPort[port] ?? "-";
  };

  const filtered = useMemo(
    () => list.filter((record) => includes(record, query)),
    [list, query]
  );

  const clients = new Set(list.map((record) => record.src_remote_addr));
  const nodes = new Set(list.map((record) => record.dst_name));

  const killConnection = async (record: ConnItem) => {
    try {
      await apiPostForm(apis.statsKillConnectionApi, { id: record.id });
      message.success("连接已断开");
      reload();
    } catch (err) {
      message.error(apiErrorMessage(err));
    }
  };

  const columns: ColumnsType<ConnItem> = [
    {
      title: "会话",
      dataIndex: "id",
      key: "id",
      width: 90,
      render: (value: number) => <span className="mono-text">#{value}</span>,
    },
    {
      title: "客户端",
      dataIndex: "src_remote_addr",
      key: "src_remote_addr",
      render: (value: string | undefined) => (
        <span className="mono-text">{value || "-"}</span>
      ),
    },
    {
      title: "入口",
      dataIndex: "src_local_addr",
      key: "src_local_addr",
      render: (value: string | undefined) => (
        <span className="mono-text">{value || "-"}</span>
      ),
    },
    {
      title: "节点",
      dataIndex: "dst_name",
      key: "dst_name",
      render: (value: string | undefined) => (
        <span className="cell-primary">{value || "-"}</span>
      ),
    },
    {
      title: "内网目标",
      key: "target",
      render: (_, record) => (
        <span className="mono-text">{targetOf(record)}</span>
      ),
    },
    {
      title: "操作",
      key: "actions",
      width: 110,
      align: "right",
      render: (_, record) => (
        <Popconfirm
          title={
            <span>
              断开该连接？
              <br />
              <span className="cell-muted">
                {record.src_remote_addr} → {record.dst_local_addr}
              </span>
            </span>
          }
          okText="断开"
          cancelText="取消"
          okButtonProps={{ danger: true }}
          onConfirm={() => killConnection(record)}
        >
          <Button
            type="text"
            size="small"
            danger
            icon={<DisconnectOutlined />}
            aria-label={`断开连接 ${record.id}`}
          />
        </Popconfirm>
      ),
    },
  ];

  const renderMobileCards = () => (
    <div className="stack">
      {filtered.map((record) => (
        <article className="conn-card" key={record.id}>
          <div className="row-between">
            <span className="mono-text">#{record.id}</span>
            <Popconfirm
              title="断开该连接？"
              okText="断开"
              cancelText="取消"
              okButtonProps={{ danger: true }}
              onConfirm={() => killConnection(record)}
            >
              <Button size="small" danger icon={<DisconnectOutlined />}>
                断开
              </Button>
            </Popconfirm>
          </div>
          <div className="divider" />
          <dl className="detail-list">
            <div className="detail-row">
              <dt>客户端</dt>
              <dd>{record.src_remote_addr}</dd>
            </div>
            <div className="detail-row">
              <dt>入口</dt>
              <dd>{record.src_local_addr}</dd>
            </div>
            <div className="detail-row">
              <dt>节点</dt>
              <dd>{record.dst_name}</dd>
            </div>
            <div className="detail-row">
              <dt>内网目标</dt>
              <dd>{targetOf(record)}</dd>
            </div>
          </dl>
        </article>
      ))}
    </div>
  );

  return (
    <>
      <PageHeader
        eyebrow="Live sessions"
        title="连接"
        description="当前正在转发的客户端会话，每 5 秒刷新一次。断开连接会立即关闭两端 socket。"
        actions={
          <Button icon={<ReloadOutlined />} onClick={reload} loading={loading}>
            刷新
          </Button>
        }
      />

      {!data && loading && <LoadingState variant="cards" />}
      {!data && error && (
        <div className="panel">
          <ErrorState error={error} onRetry={reload} title="无法读取连接列表" />
        </div>
      )}

      {data && (
        <>
          <div className="grid stats-grid" style={{ marginBottom: 20 }}>
            <StatCard
              label="活跃连接"
              value={formatNumber(list.length)}
              hint="正在转发的会话"
              icon={<LinkOutlined />}
            />
            <StatCard
              label="独立客户端"
              value={formatNumber(clients.size)}
              hint="按来源地址去重"
              icon={<ThunderboltOutlined />}
              tone="info"
            />
            <StatCard
              label="涉及节点"
              value={formatNumber(nodes.size)}
              hint="承载这些会话的 Agent"
              icon={<LinkOutlined />}
            />
          </div>

          <section className="panel">
            <div className="panel-header">
              <div>
                <h2 className="panel-title">会话列表</h2>
                <p className="panel-subtitle">
                  共 {formatNumber(list.length)} 条活跃连接
                </p>
              </div>
              <Search
                allowClear
                className="search"
                placeholder="搜索地址、端口或节点"
                prefix={<SearchOutlined />}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                aria-label="搜索连接"
              />
            </div>
            <div className="panel-body flush">
              {filtered.length === 0 ? (
                <EmptyState
                  icon={<LinkOutlined />}
                  title={
                    list.length === 0 ? "当前没有活跃连接" : "没有匹配的连接"
                  }
                  description={
                    list.length === 0
                      ? "客户端通过映射端口发起请求后，连接会显示在这里。"
                      : "尝试调整搜索关键词。"
                  }
                />
              ) : isMobile ? (
                <div className="panel-body">{renderMobileCards()}</div>
              ) : (
                <Table<ConnItem>
                  rowKey="id"
                  columns={columns}
                  dataSource={filtered}
                  pagination={false}
                  scroll={{ x: 860 }}
                />
              )}
            </div>
          </section>
        </>
      )}
    </>
  );
}
