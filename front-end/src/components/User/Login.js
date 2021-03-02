import { Form, Icon, Input, Button, Checkbox, Result } from "antd";
import { Link } from "react-router-dom";
import React from "react";
import PropTypes from "prop-types";
import "./user.css";

class NormalLoginForm extends React.Component {
  handleSubmit = e => {
    e.preventDefault();
    this.props.form.validateFields((err, values) => {
      if (!err) {
        this.props.actions.userLogin(
          values.username,
          values.password,
          values.otpcode
        );
      }
    });
  };
  handleResetError = () => {
    this.props.actions.userClearError();
  };

  render() {
    const { getFieldDecorator } = this.props.form;
    if (this.props.init === false && this.props.loading === false) {
      if (this.props.isLogin === true) {
        return (
          <Result
            status="success"
            title="登陆成功"
            extra={[
              <Button type="primary" key="console">
                <Link to="/">首页</Link>
              </Button>
            ]}
          />
        );
      } else if (this.props.isLogin === false) {
        return (
          <Result
            status="error"
            title="登陆失败"
            subTitle={this.props.error.response.data.data || ""}
            extra={[
              <Button
                type="primary"
                key="console"
                onClick={this.handleResetError}
              >
                重新登陆
              </Button>
            ]}
          />
        );
      }
    }

    return (
      <Form onSubmit={this.handleSubmit} className="login-form">
        <Form.Item>
          {getFieldDecorator("username", {
            rules: [{ required: true, message: "Please input your username!" }]
          })(
            <Input
              prefix={<Icon type="user" style={{ color: "rgba(0,0,0,.25)" }} />}
              placeholder="Username"
            />
          )}
        </Form.Item>
        <Form.Item>
          {getFieldDecorator("password", {
            rules: [{ required: true, message: "Please input your Password!" }]
          })(
            <Input
              prefix={<Icon type="lock" style={{ color: "rgba(0,0,0,.25)" }} />}
              type="password"
              placeholder="Password"
            />
          )}
        </Form.Item>
        <Form.Item>
          {getFieldDecorator("otpcode", {
            rules: [{ required: true, message: "Please input your otp code!" }]
          })(
            <Input
              prefix={<Icon type="lock" style={{ color: "rgba(0,0,0,.25)" }} />}
              type="password"
              placeholder="OTP Code"
            />
          )}
        </Form.Item>
        <Form.Item>
          {getFieldDecorator("remember", {
            valuePropName: "checked",
            initialValue: true
          })(<Checkbox>Remember me</Checkbox>)}
          {/* <a className="login-form-forgot" href="">
            Forgot password
          </a> */}
          <Button
            type="primary"
            htmlType="submit"
            className="login-form-button"
          >
            Log in
          </Button>
          {/* Or <a href="">register now!</a> */}
        </Form.Item>
      </Form>
    );
  }
}

const WrappedNormalLoginForm = Form.create({ name: "normal_login" })(
  NormalLoginForm
);
NormalLoginForm.propTypes = {
  form: PropTypes.any,
  error: PropTypes.any,
  init: PropTypes.bool,
  loading: PropTypes.bool,
  isLogin: PropTypes.bool,
  isLoginError: PropTypes.bool,
  actions: PropTypes.object.isRequired
};

export default WrappedNormalLoginForm;
