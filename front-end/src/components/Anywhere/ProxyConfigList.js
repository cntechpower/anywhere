import React, { Component } from "react";
import { Table, Spin, Result, Divider, message } from "antd";
import ButtonWithConfirm from "./ButtonWithConfirm";
import PropTypes from "prop-types";

class ProxyConfigList extends Component {
  componentDidMount() {
    this.props.actions.fetchProxyConfigs();
  }
  deleteProxyConfig = (agent_id, local_addr) => {
    const onSuccess = () => {
      message.success("配置删除成功");
      this.props.actions.fetchProxyConfigs();
    };
    const onFail = () => {
      message.error("配置删除失败");
    };
    this.props.actions.deleteProxyConfig(
      agent_id,
      local_addr,
      onSuccess,
      onFail
    );
  };

  render() {
    if (this.props.loading)
      return (
        <div className="loading">
          <Spin tip="Loading..." size="large" />
        </div>
      );
    if (this.props.error !== null && this.props.error !== undefined) {
      return (
        <Result
          status="error"
          title="网络连接失败"
          subTitle={this.props.error}
        />
      );
    }

    const columns = [
      {
        title: "AgentId",
        dataIndex: "agent_id",
        key: "AgentId"
      },
      {
        title: "公网地址",
        dataIndex: "remote_port",
        key: "RemoteAddr",
        render: text => {
          let displayText = "未知状态";
          if (text !== undefined && text !== null) {
            displayText = "0.0.0.0:" + text;
          }
          return displayText;
        }
      },
      {
        title: "内网地址",
        dataIndex: "local_addr",
        key: "LocalAddr"
      },

      {
        title: "白名单开关",
        key: "IsWhiteListOn",
        dataIndex: "is_whitelist_on",
        render: text => {
          let displayText = "未知状态";
          if (text === true) {
            displayText = "开启";
          } else {
            displayText = "关闭";
          }
          return displayText;
        }
      },
      {
        title: "白名单",
        dataIndex: "whitelist_ips",
        key: "WhiteListIps",
        ellipsis: true,
        render: text => {
          return text || "-";
        }
      },

      {
        title: "操作",
        key: "action",
        render: (text, record) => (
          <span>
            <ButtonWithConfirm
              btnName="Update"
              btnDisabled={true}
              confirmTitle="是否确认删除此配置?"
              confirmContent={
                "0.0.0.0:" + record.remote_port + " -> " + record.local_addr
              }
              fnOnOk={() =>
                this.deleteProxyConfig(record.agent_id, record.local_addr)
              }
            />
            <Divider type="vertical" />
            <ButtonWithConfirm
              btnName="Delete"
              btnDisabled={false}
              confirmTitle="是否确认删除此配置?"
              confirmContent={
                "0.0.0.0:" + record.remote_port + " -> " + record.local_addr
              }
              fnOnOk={() =>
                this.deleteProxyConfig(record.agent_id, record.local_addr)
              }
              //https://github.com/ant-design/ant-design/issues/4453
            />
          </span>
        )
      }
    ];
    return (
      <Table
        rowKey="remote_port"
        columns={columns}
        dataSource={this.props.data}
      />
    );
  }
}

ProxyConfigList.propTypes = {
  loading: PropTypes.bool.isRequired,
  error: PropTypes.string,
  data: PropTypes.array,
  actions: PropTypes.object.isRequired
};

export default ProxyConfigList;
