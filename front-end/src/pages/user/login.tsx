import { useState } from "react";
import { Alert, Button, Form, Input } from "antd";
import {
  LockOutlined,
  LoginOutlined,
  SafetyCertificateOutlined,
  UserOutlined,
} from "@ant-design/icons";
import Brand from "../../components/Brand";
import apis from "../../apis/apis";
import { apiPostForm } from "../../api/client";
import { apiErrorMessage } from "../../utils/format";
import type { LoginFormValues } from "../../types";

export default function Login() {
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (values: LoginFormValues) => {
    setSubmitting(true);
    setError(null);
    try {
      await apiPostForm(apis.userLoginApi, {
        username: values.username.trim(),
        password: values.password,
        otpcode: (values.otpcode ?? "").trim(),
      });
      try {
        window.localStorage.setItem("anywhere.user", values.username.trim());
      } catch {
        /* storage may be unavailable in private mode */
      }
      window.location.assign("/home");
    } catch (err) {
      setError(apiErrorMessage(err));
      setSubmitting(false);
    }
  };

  return (
    <div className="login-page">
      <div className="login-glow one" aria-hidden="true" />
      <div className="login-glow two" aria-hidden="true" />

      <main className="login-card">
        <div className="login-head">
          <Brand />
          <h1>登录控制台</h1>
          <p>使用用户名与密码登录；启用 OTP 的账户还需输入动态码。</p>
        </div>

        {error && (
          <Alert
            type="error"
            showIcon
            style={{ marginBottom: 18 }}
            message="登录失败"
            description={error}
          />
        )}

        <Form<LoginFormValues>
          layout="vertical"
          onFinish={handleSubmit}
          requiredMark={false}
        >
          <Form.Item
            label="用户名"
            name="username"
            rules={[{ required: true, message: "请输入用户名" }]}
          >
            <Input
              size="large"
              prefix={<UserOutlined />}
              placeholder="admin"
              autoComplete="username"
              autoFocus
              spellCheck={false}
            />
          </Form.Item>

          <Form.Item
            label="密码"
            name="password"
            rules={[{ required: true, message: "请输入密码" }]}
          >
            <Input.Password
              size="large"
              prefix={<LockOutlined />}
              placeholder="账户密码"
              autoComplete="current-password"
            />
          </Form.Item>

          <Form.Item
            label="动态码"
            name="otpcode"
            extra="启用 OTP 的账户必填，6 位数字；未启用可留空。"
          >
            <Input
              size="large"
              prefix={<SafetyCertificateOutlined />}
              placeholder="123456"
              inputMode="numeric"
              autoComplete="one-time-code"
              maxLength={6}
            />
          </Form.Item>

          <Button
            type="primary"
            htmlType="submit"
            block
            size="large"
            loading={submitting}
            icon={<LoginOutlined />}
          >
            登录
          </Button>
        </Form>

        <div className="login-foot">
          Anywhere · 内网穿透控制台
          <br />
          <code>mTLS</code> 加密传输 · 会话有效期 7 天
        </div>
      </main>
    </div>
  );
}
