import React, { Component } from "react";
import {
  Table,
  Spin,
  Result,
  Divider,
  message,
  notification,
  Form,
  Modal,
  Switch,
  Button,
  Input,
} from "antd";
import ButtonWithConfirm from "../../tools/ButtonWithConfirm";
// import Draggable from "react-draggable";
import qs from "qs";
import apis from "../../apis/apis";
import axios from "axios";

const { TextArea } = Input;

class List extends Component {
  constructor(props) {
    super(props);
    this.state = {
      loading: true,
      data: null,
      error: null,
      showUpdateModal: false,
    };
  }

  componentDidMount() {
    this.fetchProxyConfigs();
  }

  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };

  formRef = React.createRef();

  fetchProxyConfigs() {
    fetch(apis.proxyConfigListApi)
      .then((response) => response.json())
      .then(
        (response) => {
          this.setState({ loading: false, data: response });
        },
        (error) => {
          this.setState({ loading: false, error: error });
        }
      );
  }

  deleteProxyConfig = (id) => {
    const options = {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      data: qs.stringify({
        id: id,
      }),
      url: apis.proxyAgentDelApi,
    };
    return axios(options).then(
      () => {
        message.success("配置删除成功");
        this.fetchProxyConfigs();
      },
      // Note: it's important to handle errors here
      // instead of a catch() block so that we don't swallow
      // exceptions from actual bugs in components.
      (error) => {
        this.openNotificationWithIcon("error", error.message);
        console.log(error);
      }
    );
  };

  showUpdateModal = (
    user_name,
    zone_name,
    local_addr,
    remote_port,
    white_list_enable,
    white_list_ips
  ) => {
    this.setState({
      showUpdateModal: true,
      updateUserName: user_name,
      updateZoneName: zone_name,
      updateLocalAddr: local_addr,
      updateRemotePort: remote_port,
      updateWhiteListEnable: white_list_enable,
      updateWhiteListIps: white_list_ips,
    });
  };

  hideAddModal = () => {
    this.setState({
      showUpdateModal: false,
    });
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

  updateProxyConfig = (
    user_name,
    zone_name,
    local_addr,
    remote_port,
    white_list_enable,
    white_list_ips
  ) => {
    const options = {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      data: qs.stringify({
        user_name: user_name,
        zone_name: zone_name,
        local_addr: local_addr,
        remote_port: remote_port,
        white_list_enable: white_list_enable,
        white_list_ips: white_list_ips,
      }),
      url: apis.proxyConfigUpdateApi,
    };
    return axios(options).then(
      () => {
        message.success("配置修改成功");
        this.hideAddModal();
        this.fetchProxyConfigs();
      },
      // Note: it's important to handle errors here
      // instead of a catch() block so that we don't swallow
      // exceptions from actual bugs in components.
      (error) => {
        this.openNotificationWithIcon("error", error.message);
        console.log(error);
      }
    );
  };

  render() {
    if (this.state.loading)
      return (
        <div className="loading">
          <Spin tip="Loading..." size="large" />
        </div>
      );
    if (this.state.error !== null && this.state.error !== undefined) {
      return (
        <Result
          status="error"
          title="网络连接失败"
          subTitle={this.state.error.message}
        />
      );
    }

    const columns = [
      {
        title: "ID",
        dataIndex: "id",
        key: "id",
        defaultSortOrder: "descend",
        sorter: (a, b) => a.id - b.id,
      },
      {
        title: "用户",
        dataIndex: "user_name",
        key: "UserName",
      },
      {
        title: "网络区域",
        dataIndex: "zone_name",
        key: "ZoneName",
      },
      {
        title: "公网地址",
        dataIndex: "remote_port",
        key: "RemoteAddr",
        render: (text) => {
          let displayText = "未知状态";
          if (text !== undefined && text !== null) {
            displayText = "0.0.0.0:" + text;
          }
          return displayText;
        },
      },
      {
        title: "内网地址",
        dataIndex: "local_addr",
        key: "LocalAddr",
      },

      {
        title: "白名单控制",
        key: "IsWhiteListOn",
        dataIndex: "is_whitelist_on",
        render: (text) => {
          let displayText = "未知状态";
          if (text === true) {
            displayText = "开启";
          } else {
            displayText = "关闭";
          }
          return displayText;
        },
      },
      {
        title: "白名单",
        dataIndex: "whitelist_ips",
        key: "WhiteListIps",
        ellipsis: true,
        render: (text) => {
          return text || "-";
        },
      },

      {
        title: "操作",
        key: "action",
        render: (text, record) => (
          <span>
            <Button
              onClick={() =>
                this.showUpdateModal(
                  record.user_name,
                  record.zone_name,
                  record.local_addr,
                  record.remote_port,
                  record.is_whitelist_on,
                  record.whitelist_ips
                )
              }
            >
              Update
            </Button>
            <Divider type="vertical" />
            <ButtonWithConfirm
              btnName="Delete"
              btnDisabled={false}
              confirmTitle="是否确认删除此配置?"
              confirmContent={
                "0.0.0.0:" + record.remote_port + " -> " + record.local_addr
              }
              fnOnOk={() => this.deleteProxyConfig(record.id)}
              //https://github.com/ant-design/ant-design/issues/4453
            />
          </span>
        ),
      },
    ];
    return (
      <>
        <Modal
          title={
            <div
              style={{
                width: "100%",
                cursor: "move",
              }}
            >
              修改配置
            </div>
          }
          visible={this.state.showUpdateModal}
          footer={[
            <Button form="addV2rayNodeForm" key="submit" htmlType="submit">
              Submit
            </Button>,
          ]}
          onCancel={this.hideAddModal}
          // modalRender={(modal) => <Draggable>{modal}</Draggable>}
        >
          <Form
            id="addV2rayNodeForm"
            ref={this.formRef}
            layout="vertical"
            initialValues={{ modifier: "public" }}
            onFinish={(values) => {
              this.updateProxyConfig(
                this.state.updateUserName,
                this.state.updateZoneName,
                this.state.updateLocalAddr,
                this.state.updateRemotePort,
                values.white_list_enable,
                values.white_list_ips
              );
            }}
          >
            <Form.Item
              label="白名单开关"
              name="white_list_enable"
              valuePropName="checked"
              initialValue={this.state.updateWhiteListEnable}
            >
              <Switch
                checkedChildren="开"
                unCheckedChildren="关"
                onChange={(value) => {
                  this.setState({
                    updateWhiteListEnable: value,
                  });
                }}
              />
            </Form.Item>
            <Form.Item>
              <Button
                type="primary"
                disabled={!this.state.updateWhiteListEnable}
                onClick={this.addLocalIpToTextArea}
              >
                向白名单中添加本机公网地址
              </Button>
            </Form.Item>
            <Form.Item
              label="白名单地址"
              name="white_list_ips"
              initialValue={this.state.updateWhiteListIps}
            >
              <TextArea
                rows={3}
                placeholder="白名单CIDR列表, 使用逗号分隔. 例如: 180.169.60.146/32,117.186.0.0/16,223.88.249.0/24"
                disabled={!this.state.updateWhiteListEnable}
                style={{ width: "60%" }}
              />
            </Form.Item>
          </Form>
        </Modal>
        <Table
          rowKey="remote_port"
          columns={columns}
          dataSource={this.state.data}
        />
      </>
    );
  }
}

export default List;
