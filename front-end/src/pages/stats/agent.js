import React, { Component } from "react";
import { Table, Spin, Result, notification } from "antd";
// import Draggable from "react-draggable";
import apis from "../../apis/apis";

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
    this.fetchAgentList();
  }

  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };

  formRef = React.createRef();

  fetchAgentList() {
    fetch(apis.proxyAgentListApi)
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
        title: "用户名",
        dataIndex: "userName",
        key: "userName",
      },
      {
        title: "可用区",
        dataIndex: "zoneName",
        key: "zoneName",
      },
      {
        title: "节点名",
        dataIndex: "agentId",
        key: "agentId",
      },
      {
        title: "节点地址",
        dataIndex: "agentAdminAddr",
        key: "agentAdminAddr",
      },
      {
        title: "心跳发送时间",
        dataIndex: "lastAckSend",
        key: "lastAckSend",
      },
      {
        title: "心跳接收时间",
        dataIndex: "lastAckRcv",
        key: "lastAckRcv",
      },

      {
        title: "延迟",
        // dataIndex: "lastAckRcv",
        key: "lastAckRcv",
        render: (text, record) => {
          var delay =
            Date.parse(record.lastAckRcv) - Date.parse(record.lastAckSend);
          return delay + "ms";
        },
      },

      {
        title: "配置总数",
        dataIndex: "proxyConfigCount",
        key: "proxyConfigCount",
        render: (count) => {
          return count || 0;
        },
      },
    ];
    return <Table rowKey="id" columns={columns} dataSource={this.state.data} />;
  }
}

export default List;
