import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Button,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
  message,
} from "antd";
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  PlusOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import { Link } from "react-router-dom";
import PageHeader from "../../components/PageHeader";
import WhiteListFields from "../../components/WhiteListFields";
import ErrorState from "../../components/ErrorState";
import LoadingState from "../../components/LoadingState";
import useApi from "../../hooks/useApi";
import apis from "../../apis/apis";
import { apiGet, apiPostForm } from "../../api/client";
import { apiErrorMessage } from "../../utils/format";
import type {
  CreatedProxyConfig,
  ProxyFormValues,
  ZoneInfo,
} from "../../types";

interface ProtocolSelectProps {
  value?: "tcp" | "udp";
  onChange?: (value: "tcp" | "udp") => void;
}

function ProtocolSelect({ value, onChange }: ProtocolSelectProps) {
  const current = value ?? "tcp";
  return (
    <div className="segmented" role="radiogroup" aria-label="监听类型">
      {(["tcp", "udp"] as const).map((protocol) => (
        <button
          key={protocol}
          type="button"
          role="radio"
          aria-checked={current === protocol}
          onClick={() => onChange?.(protocol)}
        >
          {protocol}
        </button>
      ))}
    </div>
  );
}

export default function AddProxyConfig() {
  const [form] = Form.useForm<ProxyFormValues>();
  const [submitting, setSubmitting] = useState(false);
  const [created, setCreated] = useState<CreatedProxyConfig | null>(null);

  const zonesApi = useApi<ZoneInfo[]>(
    (signal) => apiGet<ZoneInfo[]>(apis.proxyZoneListApi, { signal }),
    { pollMs: 30000 }
  );

  const zones = useMemo(
    () => (Array.isArray(zonesApi.data) ? zonesApi.data : []),
    [zonesApi.data]
  );
  const zoneOptions = zones.map((zone) => ({
    value: `${zone.user_name}&&&${zone.zone_name}`,
    label: `${zone.user_name} / ${zone.zone_name}`,
  }));

  useEffect(() => {
    // Convenience for the common single-zone deployment.
    if (zones.length === 1 && !form.getFieldValue("zone")) {
      form.setFieldsValue({
        zone: `${zones[0].user_name}&&&${zones[0].zone_name}`,
      });
    }
  }, [zones, form]);

  const handleSubmit = async (values: ProxyFormValues) => {
    const [userName, zoneName] = String(values.zone).split("&&&");
    setSubmitting(true);
    try {
      await apiPostForm(apis.proxyConfigAddApi, {
        user_name: userName,
        zone_name: zoneName,
        remote_port: values.remote_port,
        local_addr: `${values.local_ip}:${values.local_port}`,
        white_list_enable: Boolean(values.white_list_enable),
        white_list_ips: values.white_list_ips ?? "",
        listen_type: values.listen_type ?? "tcp",
      });
      message.success("代理配置已创建");
      setCreated({
        userName,
        zoneName,
        remotePort: values.remote_port,
        localAddr: `${values.local_ip}:${values.local_port}`,
        listenType: values.listen_type ?? "tcp",
        whiteList: Boolean(values.white_list_enable),
      });
      form.resetFields();
    } catch (error) {
      message.error(apiErrorMessage(error));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <PageHeader
        eyebrow="New proxy"
        title="新增配置"
        description="把服务端的一个公网端口映射到 Agent 所在内网的服务地址。创建后立即生效。"
        actions={
          <Link to="/configs/list">
            <Button icon={<ArrowLeftOutlined />}>返回列表</Button>
          </Link>
        }
      />

      {zonesApi.error && !zonesApi.data && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <ErrorState
            error={zonesApi.error}
            onRetry={zonesApi.reload}
            title="无法读取可用区"
          />
        </div>
      )}

      {zonesApi.loading && !zonesApi.data && (
        <div className="panel" style={{ marginBottom: 16 }}>
          <div className="panel-body">
            <LoadingState variant="panels" />
          </div>
        </div>
      )}

      {zonesApi.data && zones.length === 0 && (
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
          message="暂无可用节点"
          description="请先在 Agent 端启动并连接到服务端，可用区出现后才能创建代理配置。"
          action={
            <Link to="/stats/agents">
              <Button size="small">查看节点</Button>
            </Link>
          }
        />
      )}

      {created && (
        <section className="panel" style={{ marginBottom: 20 }}>
          <div className="panel-body">
            <div className="row-between" style={{ alignItems: "flex-start" }}>
              <div>
                <span className="eyebrow">
                  <CheckCircleOutlined /> Created
                </span>
                <h2 className="panel-title" style={{ marginTop: 8 }}>
                  配置创建成功
                </h2>
                <p className="panel-subtitle">
                  {created.listenType.toUpperCase()} · 0.0.0.0:
                  {created.remotePort} → {created.localAddr}
                </p>
              </div>
              <div className="inline-actions">
                <Button onClick={() => setCreated(null)}>继续添加</Button>
                <Link to="/configs/list">
                  <Button type="primary">查看配置</Button>
                </Link>
              </div>
            </div>
            <div className="divider" />
            <dl className="detail-list">
              <div className="detail-row">
                <dt>可用区</dt>
                <dd>
                  {created.userName} / {created.zoneName}
                </dd>
              </div>
              <div className="detail-row">
                <dt>白名单</dt>
                <dd>{created.whiteList ? "已开启" : "未开启"}</dd>
              </div>
            </dl>
          </div>
        </section>
      )}

      <Form<ProxyFormValues>
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{
          listen_type: "tcp",
          local_port: 22,
          white_list_enable: false,
        }}
      >
        <div className="grid two-col" style={{ alignItems: "start" }}>
          <section className="panel">
            <div className="panel-header">
              <div>
                <h2 className="panel-title">路由与目标</h2>
                <p className="panel-subtitle">公网入口与内网服务的映射</p>
              </div>
            </div>
            <div className="panel-body">
              <Form.Item
                label="可用区"
                name="zone"
                rules={[{ required: true, message: "请选择可用区" }]}
                extra={
                  zones.length === 0
                    ? "暂无可用区，请先启动 Agent。"
                    : "可用区由已连接的 Agent 自动注册。"
                }
              >
                <Select
                  showSearch
                  optionFilterProp="label"
                  placeholder="选择用户 / 可用区"
                  options={zoneOptions}
                  loading={zonesApi.loading}
                  notFoundContent="没有可用区"
                />
              </Form.Item>

              <Form.Item label="监听类型" name="listen_type">
                <ProtocolSelect />
              </Form.Item>

              <Form.Item
                label="对外端口"
                name="remote_port"
                rules={[
                  { required: true, message: "请输入对外端口" },
                  {
                    type: "number",
                    min: 1,
                    max: 65535,
                    message: "端口范围 1 - 65535",
                  },
                ]}
                extra="客户端通过 server:端口 访问该隧道。"
              >
                <InputNumber
                  min={1}
                  max={65535}
                  style={{ width: "100%" }}
                  placeholder="8080"
                />
              </Form.Item>

              <div className="grid form-grid">
                <Form.Item
                  label="内网地址"
                  name="local_ip"
                  rules={[
                    { required: true, message: "请输入内网地址" },
                    {
                      pattern: /^[a-zA-Z0-9.-]+$/,
                      message: "请输入 IPv4 或主机名（不含端口）",
                    },
                  ]}
                  extra="Agent 侧可访问的地址。"
                >
                  <Input placeholder="10.0.0.12" spellCheck={false} />
                </Form.Item>

                <Form.Item
                  label="内网端口"
                  name="local_port"
                  rules={[
                    { required: true, message: "请输入内网端口" },
                    {
                      type: "number",
                      min: 1,
                      max: 65535,
                      message: "端口范围 1 - 65535",
                    },
                  ]}
                  extra="内网服务监听的端口。"
                >
                  <InputNumber
                    min={1}
                    max={65535}
                    style={{ width: "100%" }}
                    placeholder="22"
                  />
                </Form.Item>
              </div>
            </div>
          </section>

          <section className="panel">
            <div className="panel-header">
              <div>
                <h2 className="panel-title">访问控制</h2>
                <p className="panel-subtitle">按来源 IP 限制可访问的客户端</p>
              </div>
            </div>
            <div className="panel-body">
              <Form.Item
                label="白名单开关"
                name="white_list_enable"
                valuePropName="checked"
                extra="关闭时允许所有来源访问；开启后仅放行下列 CIDR。"
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

              <Alert
                type="info"
                showIcon
                message="不会自动放行内网网段"
                description="开启白名单后，请显式添加需要放行的 10/8、172.16/12、192.168/16 等网段。"
              />
            </div>
          </section>
        </div>

        <div className="page-actions" style={{ marginTop: 20 }}>
          <Button
            type="primary"
            htmlType="submit"
            icon={<PlusOutlined />}
            loading={submitting}
            disabled={zones.length === 0}
          >
            创建配置
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              form.resetFields();
              setCreated(null);
            }}
          >
            重置
          </Button>
        </div>
      </Form>
    </>
  );
}
