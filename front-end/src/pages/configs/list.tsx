import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Button,
  Form,
  Input,
  Modal,
  Popconfirm,
  Switch,
  Table,
  Tooltip,
  message,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from "@ant-design/icons";
import { Link } from "react-router-dom";
import PageHeader from "../../components/PageHeader";
import StatusBadge from "../../components/StatusBadge";
import EmptyState from "../../components/EmptyState";
import ErrorState from "../../components/ErrorState";
import LoadingState from "../../components/LoadingState";
import WhiteListFields from "../../components/WhiteListFields";
import useApi from "../../hooks/useApi";
import useMediaQuery from "../../hooks/useMediaQuery";
import apis from "../../apis/apis";
import { apiGet, apiPostForm } from "../../api/client";
import {
  apiErrorMessage,
  formatBytes,
  formatNumber,
} from "../../utils/format";
import type { ProxyConfig, WhiteListValues } from "../../types";

const { Search } = Input;

function matches(record: ProxyConfig, query: string): boolean {
  if (!query) {
    return true;
  }
  const haystack = [
    record.user_name,
    record.zone_name,
    record.local_addr,
    record.remote_port,
    record.whitelist_ips,
    record.listen_type,
    record.id,
  ]
    .filter((value) => value !== undefined && value !== null)
    .join(" ")
    .toLowerCase();
  return haystack.includes(query.toLowerCase());
}

function listenType(record: ProxyConfig): "TCP" | "UDP" {
  return (record.listen_type ?? "tcp").toLowerCase() === "udp" ? "UDP" : "TCP";
}

function whitelistOn(record: ProxyConfig): boolean {
  return Boolean(record.is_whitelist_on);
}

export default function ProxyConfigList() {
  const [query, setQuery] = useState("");
  const [editing, setEditing] = useState<ProxyConfig | null>(null);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm<WhiteListValues>();
  const isMobile = useMediaQuery("(max-width: 860px)");

  const { data, error, loading, reload } = useApi<ProxyConfig[]>(
    (signal) => apiGet<ProxyConfig[]>(apis.proxyConfigListApi, { signal }),
    { pollMs: 15000 }
  );

  // Keep the array identity stable so downstream useMemo hooks are not
  // invalidated on every render.
  const list = useMemo(() => (Array.isArray(data) ? data : []), [data]);
  const filtered = useMemo(
    () =>
      list
        .filter((record) => matches(record, query))
        .sort((a, b) => Number(a.remote_port) - Number(b.remote_port)),
    [list, query]
  );

  const whitelistCount = list.filter(whitelistOn).length;
  const udpCount = list.filter((record) => listenType(record) === "UDP").length;

  useEffect(() => {
    if (editing) {
      form.setFieldsValue({
        white_list_enable: whitelistOn(editing),
        white_list_ips: editing.whitelist_ips ?? "",
      });
    }
  }, [editing, form]);

  const handleDelete = async (record: ProxyConfig) => {
    try {
      await apiPostForm(apis.proxyAgentDelApi, { id: record.id });
      message.success("配置已删除");
      reload();
    } catch (err) {
      message.error(apiErrorMessage(err));
    }
  };

  const handleUpdate = async (values: WhiteListValues) => {
    if (!editing) {
      return;
    }
    setSaving(true);
    try {
      await apiPostForm(apis.proxyConfigUpdateApi, {
        user_name: editing.user_name,
        zone_name: editing.zone_name,
        local_addr: editing.local_addr,
        remote_port: editing.remote_port,
        white_list_enable: Boolean(values.white_list_enable),
        white_list_ips: values.white_list_ips ?? "",
      });
      message.success("白名单配置已更新");
      setEditing(null);
      reload();
    } catch (err) {
      message.error(apiErrorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  const columns: ColumnsType<ProxyConfig> = [
    {
      title: "公网入口",
      key: "entry",
      width: 170,
      sorter: (a, b) => Number(a.remote_port) - Number(b.remote_port),
      render: (_, record) => (
        <div className="cell-stack">
          <span className="tag-mono">{listenType(record)}</span>
          <span className="mono-text">:{record.remote_port}</span>
        </div>
      ),
    },
    {
      title: "内网目标",
      dataIndex: "local_addr",
      key: "local_addr",
      render: (value: string | undefined) => (
        <span className="mono-text">{value || "-"}</span>
      ),
    },
    {
      title: "归属",
      key: "owner",
      render: (_, record) => (
        <div className="cell-stack">
          <span className="cell-primary">{record.user_name || "-"}</span>
          <span className="cell-muted">{record.zone_name || "-"}</span>
        </div>
      ),
    },
    {
      title: "流量（进 / 出）",
      key: "flow",
      width: 190,
      render: (_, record) => (
        <span className="mono-text">
          {formatBytes(record.network_flow_remote_to_local_in_bytes)} /{" "}
          {formatBytes(record.network_flow_local_to_remote_in_bytes)}
        </span>
      ),
    },
    {
      title: "连接 / 拒绝",
      key: "connects",
      width: 130,
      render: (_, record) => (
        <span className="mono-text">
          {formatNumber(record.proxy_connect_count)} /{" "}
          <span
            style={{
              color:
                Number(record.proxy_connect_reject_count) > 0
                  ? "var(--danger)"
                  : "inherit",
            }}
          >
            {formatNumber(record.proxy_connect_reject_count)}
          </span>
        </span>
      ),
    },
    {
      title: "白名单",
      key: "whitelist",
      width: 130,
      render: (_, record) =>
        whitelistOn(record) ? (
          <Tooltip title={record.whitelist_ips || "未配置 CIDR"}>
            <StatusBadge tone="ok">已开启</StatusBadge>
          </Tooltip>
        ) : (
          <StatusBadge>未开启</StatusBadge>
        ),
    },
    {
      title: "操作",
      key: "actions",
      width: 110,
      align: "right",
      render: (_, record) => (
        <div className="inline-actions" style={{ justifyContent: "flex-end" }}>
          <Tooltip title="编辑白名单">
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              aria-label={`编辑端口 ${record.remote_port} 的白名单`}
              onClick={() => setEditing(record)}
            />
          </Tooltip>
          <Popconfirm
            title={
              <span>
                删除该配置？
                <br />
                <span className="cell-muted">
                  0.0.0.0:{record.remote_port} → {record.local_addr}
                </span>
              </span>
            }
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
            onConfirm={() => handleDelete(record)}
          >
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
              aria-label={`删除端口 ${record.remote_port} 的配置`}
            />
          </Popconfirm>
        </div>
      ),
    },
  ];

  const renderMobileCards = () => (
    <div className="stack">
      {filtered.map((record) => (
        <article className="config-card" key={record.id}>
          <div className="row-between">
            <div className="cell-stack">
              <span className="tag-mono">{listenType(record)}</span>
              <span className="mono-text" style={{ fontSize: 16 }}>
                :{record.remote_port}
              </span>
            </div>
            <div className="inline-actions">
              <Button
                size="small"
                icon={<EditOutlined />}
                onClick={() => setEditing(record)}
                aria-label="编辑白名单"
              >
                编辑
              </Button>
              <Popconfirm
                title="删除该配置？"
                okText="删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
                onConfirm={() => handleDelete(record)}
              >
                <Button
                  size="small"
                  danger
                  icon={<DeleteOutlined />}
                  aria-label="删除配置"
                />
              </Popconfirm>
            </div>
          </div>
          <div className="divider" />
          <dl className="detail-list">
            <div className="detail-row">
              <dt>内网目标</dt>
              <dd>{record.local_addr || "-"}</dd>
            </div>
            <div className="detail-row">
              <dt>归属</dt>
              <dd>
                {record.user_name} / {record.zone_name}
              </dd>
            </div>
            <div className="detail-row">
              <dt>流量（进 / 出）</dt>
              <dd>
                {formatBytes(record.network_flow_remote_to_local_in_bytes)} /{" "}
                {formatBytes(record.network_flow_local_to_remote_in_bytes)}
              </dd>
            </div>
            <div className="detail-row">
              <dt>连接 / 拒绝</dt>
              <dd>
                {formatNumber(record.proxy_connect_count)} /{" "}
                {formatNumber(record.proxy_connect_reject_count)}
              </dd>
            </div>
            <div className="detail-row">
              <dt>白名单</dt>
              <dd>{whitelistOn(record) ? "已开启" : "未开启"}</dd>
            </div>
          </dl>
        </article>
      ))}
    </div>
  );

  return (
    <>
      <PageHeader
        eyebrow="Proxy configs"
        title="代理配置"
        description="对外端口到内网目标的映射。修改白名单后立即生效，无需重启 Agent。"
        actions={
          <>
            <Button icon={<ReloadOutlined />} onClick={reload} loading={loading}>
              刷新
            </Button>
            <Link to="/configs/add">
              <Button type="primary" icon={<PlusOutlined />}>
                新增配置
              </Button>
            </Link>
          </>
        }
      />

      {error && data && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          message="刷新失败，正在显示上一次成功加载的数据"
          description={apiErrorMessage(error)}
        />
      )}

      {!data && loading && (
        <div className="panel">
          <div className="panel-body">
            <LoadingState rows={6} />
          </div>
        </div>
      )}

      {!data && error && (
        <div className="panel">
          <ErrorState error={error} onRetry={reload} title="无法读取配置列表" />
        </div>
      )}

      {data && (
        <section className="panel">
          <div className="panel-header">
            <div>
              <h2 className="panel-title">配置列表</h2>
              <p className="panel-subtitle">
                共 {formatNumber(list.length)} 条 ·{" "}
                {formatNumber(udpCount)} 条 UDP ·{" "}
                {formatNumber(whitelistCount)} 条启用白名单
              </p>
            </div>
            <div className="toolbar">
              <Search
                allowClear
                className="search"
                placeholder="搜索端口、地址、用户或可用区"
                prefix={<SearchOutlined />}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                aria-label="搜索代理配置"
              />
            </div>
          </div>

          <div className="panel-body flush">
            {filtered.length === 0 ? (
              <EmptyState
                icon={<SearchOutlined />}
                title={list.length === 0 ? "还没有代理配置" : "没有匹配的配置"}
                description={
                  list.length === 0
                    ? "创建第一条配置，把公网端口映射到内网服务。"
                    : "尝试调整搜索关键词。"
                }
                action={
                  list.length === 0 ? (
                    <Link to="/configs/add">
                      <Button type="primary" icon={<PlusOutlined />}>
                        新建配置
                      </Button>
                    </Link>
                  ) : null
                }
              />
            ) : isMobile ? (
              <div className="panel-body">{renderMobileCards()}</div>
            ) : (
              <Table<ProxyConfig>
                rowKey="id"
                columns={columns}
                dataSource={filtered}
                scroll={{ x: 980 }}
                pagination={{
                  pageSize: 10,
                  showSizeChanger: true,
                  pageSizeOptions: ["10", "20", "50"],
                  showTotal: (total) => `共 ${total} 条`,
                }}
              />
            )}
          </div>
        </section>
      )}

      <Modal
        title="编辑白名单"
        open={Boolean(editing)}
        onCancel={() => setEditing(null)}
        onOk={() => form.submit()}
        okText="保存"
        cancelText="取消"
        confirmLoading={saving}
        destroyOnClose
      >
        {editing && (
          <>
            <p className="cell-muted" style={{ marginTop: 0 }}>
              {listenType(editing)} · 0.0.0.0:{editing.remote_port} →{" "}
              {editing.local_addr}
            </p>
            <Form<WhiteListValues>
              form={form}
              layout="vertical"
              onFinish={handleUpdate}
              initialValues={{
                white_list_enable: whitelistOn(editing),
                white_list_ips: editing.whitelist_ips ?? "",
              }}
            >
              <Form.Item
                label="白名单开关"
                name="white_list_enable"
                valuePropName="checked"
              >
                <Switch checkedChildren="开" unCheckedChildren="关" />
              </Form.Item>
              <Form.Item noStyle shouldUpdate>
                {({ getFieldValue }) => (
                  <WhiteListFields
                    form={form}
                    enabled={Boolean(getFieldValue("white_list_enable"))}
                  />
                )}
              </Form.Item>
            </Form>
          </>
        )}
      </Modal>
    </>
  );
}
