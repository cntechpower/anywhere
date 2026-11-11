import { useState } from "react";
import { Button, Form, Input, message } from "antd";
import type { FormInstance } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import apis from "../apis/apis";
import { apiGet } from "../api/client";
import { apiErrorMessage } from "../utils/format";

interface WhiteListForm {
  white_list_ips?: string;
}

interface WhiteListFieldsProps<T extends WhiteListForm> {
  form: FormInstance<T>;
  enabled: boolean;
}

export default function WhiteListFields<T extends WhiteListForm>({
  form,
  enabled,
}: WhiteListFieldsProps<T>) {
  const [loading, setLoading] = useState(false);

  const addLocalIp = async () => {
    setLoading(true);
    try {
      const ip = await apiGet<string>(apis.getLocalIpAPI);
      const current = (form.getFieldValue("white_list_ips") ?? "").trim();
      const cleaned = current.replace(/,\s*$/, "");
      const next = cleaned ? `${cleaned},${ip}/32` : `${ip}/32`;
      form.setFieldValue("white_list_ips", next);
    } catch (error) {
      message.error(apiErrorMessage(error));
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Form.Item
        label="白名单 CIDR"
        name="white_list_ips"
        extra="逗号分隔的 CIDR 列表，例如 203.0.113.10/32,198.51.100.0/24。"
      >
        <Input.TextArea
          rows={3}
          disabled={!enabled}
          spellCheck={false}
          placeholder="203.0.113.10/32,198.51.100.0/24"
        />
      </Form.Item>
      <div style={{ marginTop: -8, marginBottom: 18 }}>
        <Button
          size="small"
          icon={<PlusOutlined />}
          loading={loading}
          disabled={!enabled}
          onClick={addLocalIp}
        >
          添加本机公网 IP
        </Button>
      </div>
    </>
  );
}
