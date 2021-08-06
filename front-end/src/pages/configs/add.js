import {
  Form,
  Input,
  Button,
  Select,
  InputNumber,
  Switch,
  notification,
} from "antd";
import React from "react";
import qs from "qs";
import apis from "../../apis/apis";
import axios from "axios";

const { Option } = Select;
const { TextArea } = Input;

class Add extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      zoneLoading: true,
      whiteListEnabled: true,
      zones: [],
    };
  }

  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };

  formRef = React.createRef();

  componentDidMount() {
    this.fetchAgentLists();
  }

  fetchAgentLists() {
    fetch(apis.proxyZoneListApi)
      .then((response) => response.json())
      .then(
        (response) => {
          this.setState({
            zoneLoading: false,
            zones: response,
          });
        },
        (error) => {
          this.openNotificationWithIcon("error", error.message);
          console.log(error);
        }
      );
  }

  handleSubmit = (values) => {
    console.log(values.zone);
    const zones = values.zone.split("&&&");
    const options = {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      data: qs.stringify({
        user_name: zones[0],
        zone_name: zones[1],
        remote_port: values.remote_port,
        local_addr: values.local_ip + ":" + values.local_port,
        white_list_enable: values.white_list_enable,
        white_list_ips: values.white_list_ips || "",
        listen_type: values.listen_type,
      }),
      url: apis.proxyConfigAddApi,
    };
    axios(options).then(
      () => {
        this.openNotificationWithIcon("success", "添加成功");
        this.formRef.current.resetFields();
      },
      (error) => {
        this.openNotificationWithIcon("error", error.message);
        console.log(error);
      }
    );
  };

  addLocalIpToTextAreaInner = (localIp) => {
    var oldCidrs = this.formRef.current.getFieldValue("white_list_ips") || "";
    if (oldCidrs !== "") {
      oldCidrs = oldCidrs + ",";
    }
    this.formRef.current.setFieldsValue({
      white_list_ips: oldCidrs + localIp + "/32",
    });
  };

  addLocalIpToTextArea = () => {
    console.log(this.formRef);
    fetch(apis.getLocalIpAPI)
      .then((response) => response.json())
      .then(
        (response) => {
          this.addLocalIpToTextAreaInner(response);
        },
        (error) => {
          this.openNotificationWithIcon("error", error.message);
          console.log(error);
        }
      );
  };

  handleSelectChange = (value) => {
    this.setState({
      whiteListEnabled: value,
    });
  };

  render() {
    const zones = this.state.zones.map((d) => (
      <Option key={d.user_name + "&&&" + d.zone_name}>
        {d.user_name + "--" + d.zone_name}
      </Option>
    ));
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
      <Form
        {...layout}
        onFinish={this.handleSubmit}
        className="proxy-config-add-form"
        ref={this.formRef}
      >
        <Form.Item
          label="网络区域"
          name="zone"
          rules={[{ required: true, message: "Please select a zone!" }]}
        >
          <Select
            showSearch
            style={{ width: 200 }}
            placeholder="Select a zone"
            optionFilterProp="children"
            notFoundContent="No zone Found"
            disabled={this.props.zoneLoading}
          >
            {zones}
          </Select>
        </Form.Item>
        <Form.Item label="监听类型" name="listen_type" initialValue="tcp">
          <Select defaultValue="tcp" style={{ width: 200 }}>
            <Option value="tcp">TCP</Option>
            <Option value="udp">UDP</Option>
          </Select>
        </Form.Item>
        <Form.Item
          label="外部监听端口"
          name="remote_port"
          rules={[{ required: true, message: "Please input a Port!" }]}
        >
          <Input addonBefore="0.0.0.0:" style={{ width: "30%" }} />
        </Form.Item>
        <Form.Item
          label="内网地址"
          name="local_ip"
          rules={[{ required: true, message: "UPlease input a IP!" }]}
        >
          <Input style={{ width: "50%" }} />
        </Form.Item>
        <Form.Item
          label="内部映射端口"
          name="local_port"
          initialValue="22"
          rules={[{ required: true, message: "UPlease input a IP!" }]}
        >
          <InputNumber />
        </Form.Item>

        <Form.Item
          label="白名单开关"
          name="white_list_enable"
          valuePropName="checked"
          initialValue="true"
        >
          <Switch
            checkedChildren="开"
            unCheckedChildren="关"
            onChange={this.handleSelectChange}
          />
        </Form.Item>
        <Form.Item {...tailLayout}>
          <Button
            type="primary"
            disabled={!this.state.whiteListEnabled}
            onClick={this.addLocalIpToTextArea}
          >
            向白名单中添加本机公网地址
          </Button>
        </Form.Item>
        <Form.Item label="白名单地址" name="white_list_ips">
          <TextArea
            rows={3}
            placeholder="白名单CIDR列表, 使用逗号分隔. 例如: 180.169.60.146/32,117.186.0.0/16,223.88.249.0/24"
            disabled={!this.state.whiteListEnabled}
            style={{ width: "60%" }}
          />
        </Form.Item>
        <Form.Item {...tailLayout}>
          <Button type="primary" htmlType="submit">
            添加
          </Button>
        </Form.Item>
      </Form>
    );
  }
}

export default Add;
