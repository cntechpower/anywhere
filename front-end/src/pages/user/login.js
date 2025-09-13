import { Form, Input, Button, notification, Checkbox } from "antd";
import { Redirect } from "react-router";
import React from "react";
import "./css/login.css";
import qs from "qs";
import apis from "../../apis/apis";
import axios from "axios";
import { UserDeleteOutlined, LockOutlined } from "@ant-design/icons";

class Login extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      isLogin: false,
    };
  }

  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };

  handleSubmit = (values) => {
    console.log(values);
    const options = {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      data: qs.stringify({
        username: values.username,
        otpcode: values.otpcode,
      }),
      url: apis.userLoginApi,
    };
    axios(options).then(
      () => {
        this.openNotificationWithIcon("success", "登录成功");
        this.setState({
          isLogin: true,
        });
      },
      (error) => {
        this.openNotificationWithIcon("error", error.message);
        console.log(error);
      }
    );
  };

  render() {
    if (this.state.isLogin === true) {
      return <Redirect to="/" />;
    }
    const layout = {
      labelCol: {
        span: 8,
      },
      wrapperCol: {
        span: 16,
      },
    };

    const tailLayout = {
      wrapperCol: {
        offset: 8,
        span: 16,
      },
    };

    return (
      <Form {...layout} onFinish={this.handleSubmit} className="login-form">
        <Form.Item
          label="用户名"
          name="username"
          rules={[{ required: true, message: "Username is required" }]}
        >
          <Input
            prefix={
              <UserDeleteOutlined
                type="user"
                style={{ color: "rgba(0,0,0,.25)" }}
              />
            }
            placeholder="Username"
          />
        </Form.Item>
        <Form.Item
          label="动态码"
          name="otpcode"
          rules={[{ required: true, message: "OtpCode is required" }]}
        >
          <Input
            prefix={
              <LockOutlined type="lock" style={{ color: "rgba(0,0,0,.25)" }} />
            }
            type="password"
            placeholder="OTP Code"
          />
        </Form.Item>

        <Form.Item {...tailLayout}>
          <Button type="primary" htmlType="submit">
            Submit
          </Button>
        </Form.Item>
      </Form>
    );
  }
}

export default Login;
