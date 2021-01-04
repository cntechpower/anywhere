import { Form, Input, Button, Select, InputNumber, Switch, Result } from "antd";
import { Link } from "react-router-dom";
import React from "react";
import PropTypes from "prop-types";

const { Option } = Select;
const { TextArea } = Input;

class ProxyConfigAddForm extends React.Component {
  componentDidMount() {
    this.props.actions.fetchAgentLists();
  }
  state = {
    whiteListEnabled: false
  };
  handleSubmit = e => {
    e.preventDefault();
    this.props.form.validateFields((err, values) => {
      if (!err) {
        this.props.actions.postProxyConfig(values);
      }
    });
  };

  handleResetError = () => {
    this.props.actions.clearErrorState();
  };

  addLocalIpToTextAreaInner = localIp => {
    var oldCidrs = this.props.form.getFieldValue("white_list_ips") || "";
    if (oldCidrs !== "") {
      oldCidrs = oldCidrs + ",";
    }
    this.props.form.setFieldsValue({
      white_list_ips: oldCidrs + localIp + "/32"
    });
  };

  addLocalIpToTextArea = () => {
    this.props.actions.fetchGetLocalIp(this.addLocalIpToTextAreaInner);
  };

  handleSelectChange = value => {
    this.setState({
      whiteListEnabled: value
    });
  };
  render() {
    const { getFieldDecorator } = this.props.form;
    if (this.props.init === false && this.props.creating === false) {
      if (this.props.createdOk === true) {
        return (
          <Result
            status="success"
            title="添加配置成功"
            extra={[
              <Button
                type="primary"
                key="console"
                onClick={this.handleResetError}
              >
                <Link to="/proxy/list">配置列表</Link>
              </Button>,
              <Button
                type="primary"
                key="addMore"
                onClick={this.handleResetError}
              >
                再次添加
              </Button>
            ]}
          />
        );
      } else if (this.props.createdOk === false) {
        return (
          <Result
            status="error"
            title="添加配置失败"
            subTitle={this.props.error.response.data}
            extra={[
              <Button
                type="primary"
                key="console"
                onClick={this.handleResetError}
              >
                <Link to="/proxy/list">配置列表</Link>
              </Button>,
              <Button
                type="primary"
                key="console1"
                onClick={this.handleResetError}
              >
                重新添加
              </Button>
            ]}
          />
        );
      }
    }

    const options = this.props.agents.map(d => (
      <Option key={d.agentId}>{d.agentId}</Option>
    ));

    return (
      <Form onSubmit={this.handleSubmit} className="proxy-config-add-form">
        <Form.Item label="Agent">
          {getFieldDecorator("agent_id", {
            rules: [{ required: true, message: "Please select a agent!" }]
          })(
            <Select
              showSearch
              style={{ width: 200 }}
              placeholder="Select a agent"
              optionFilterProp="children"
              notFoundContent="No agent Found"
              disabled={this.props.agentsLoading}
            >
              {options}
            </Select>
          )}
        </Form.Item>
        <Form.Item label="外部监听端口">
          {getFieldDecorator("remote_port", {
            initialValue: "22",
            rules: [
              {
                required: true,
                message: "Please input a Port!"
              }
            ]
          })(<Input addonBefore="0.0.0.0:" />)}
        </Form.Item>
        <Form.Item label="内网地址">
          {getFieldDecorator("local_ip", {
            initialValue: "127.0.0.1",
            rules: [
              {
                required: true,
                message: "Please input a IP!"
              }
            ]
          })(<Input />)}
        </Form.Item>
        <Form.Item label="内部映射端口">
          {getFieldDecorator("local_port", {
            initialValue: "22",
            rules: [
              {
                required: true,
                message: "Please input a Port!"
              }
            ]
          })(<InputNumber />)}
        </Form.Item>

        <Form.Item label="白名单开关">
          {getFieldDecorator("white_list_enable", {
            valuePropName: "checked",
            initialValue: false
          })(
            <Switch
              checkedChildren="开"
              unCheckedChildren="关"
              onChange={this.handleSelectChange}
            />
          )}
        </Form.Item>
        <Button
          type="primary"
          icon="diff"
          disabled={!this.state.whiteListEnabled}
          onClick={this.addLocalIpToTextArea}
        >
          向白名单中添加本机公网地址
        </Button>
        <Form.Item label="白名单地址">
          {getFieldDecorator(
            "white_list_ips",
            {}
          )(
            <TextArea
              rows={3}
              placeholder="白名单CIDR列表, 使用逗号分隔. 例如: 180.169.60.146/32,117.186.0.0/16,223.88.249.0/24"
              disabled={!this.state.whiteListEnabled}
            />
          )}
        </Form.Item>
        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            className="login-form-button"
            loading={this.props.creating}
          >
            添加
          </Button>
        </Form.Item>
      </Form>
    );
  }
}
ProxyConfigAddForm.propTypes = {
  init: PropTypes.bool,
  creating: PropTypes.bool,
  createdOk: PropTypes.bool,
  error: PropTypes.any,
  actions: PropTypes.object.isRequired,
  config: PropTypes.object,
  form: PropTypes.any,
  agentsLoading: PropTypes.bool,
  agents: PropTypes.any,
  localIp: PropTypes.string
};

const WrappedProxyConfigAddForm = Form.create({
  name: "proxy-config-add-form"
})(ProxyConfigAddForm);

export default WrappedProxyConfigAddForm;
