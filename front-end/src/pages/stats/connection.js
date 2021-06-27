import React, { Component } from "react";
import { Table, Spin, Result, Divider, message, notification } from "antd";
import ButtonWithConfirm from "../../tools/ButtonWithConfirm";
// import Draggable from "react-draggable";
import qs from "qs";
import apis from "../../apis/apis";
import axios from "axios";

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
    this.fetchConnectionList();
  }

  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };

  formRef = React.createRef();

  fetchConnectionList() {
    fetch(apis.statsGetConnectionListApi)
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

  killConnection = (id) => {
    const options = {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      data: qs.stringify({
        id: id,
      }),
      url: apis.statsKillConnectionApi,
    };
    return axios(options).then(
      () => {
        message.success("连接Kill成功");
        this.fetchConnectionList();
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
        title: "远程地址",
        dataIndex: "src_remote_addr",
        key: "src_remote_addr",
        // specify the condition of filtering result
        // here is that finding the name started with `value`
        onFilter: (value, record) => record.name.indexOf(value) === 0,
        sorter: (a, b) => a.src_remote_addr.length - b.src_remote_addr.length,
        sortDirections: ["descend"],
      },
      {
        title: "映射端口",
        dataIndex: "src_local_addr",
        key: "src_local_addr",
        // specify the condition of filtering result
        // here is that finding the name started with `value`
        onFilter: (value, record) => record.name.indexOf(value) === 0,
        sorter: (a, b) => a.src_local_addr.length - b.src_local_addr.length,
        sortDirections: ["descend"],
      },
      {
        title: "内网节点",
        dataIndex: "dst_name",
        key: "dst_name",
        // specify the condition of filtering result
        // here is that finding the name started with `value`
        onFilter: (value, record) => record.name.indexOf(value) === 0,
        sorter: (a, b) => a.dst_name.length - b.dst_name.length,
        sortDirections: ["descend"],
      },
      {
        title: "内网地址",
        dataIndex: "src_name",
        key: "src_name",
        // specify the condition of filtering result
        // here is that finding the name started with `value`
        onFilter: (value, record) => record.name.indexOf(value) === 0,
        sorter: (a, b) => a.src_name.length - b.src_name.length,
        sortDirections: ["descend"],
      },

      {
        title: "操作",
        key: "action",
        render: (text, record) => (
          <span>
            <Divider type="vertical" />
            <ButtonWithConfirm
              btnName="Kill"
              btnDisabled={false}
              confirmTitle="是否确认Kill此连接?"
              confirmContent={record.src_remote_addr + " -> " + record.src_name}
              fnOnOk={() => this.killConnection(record.id)}
              //https://github.com/ant-design/ant-design/issues/4453
            />
          </span>
        ),
      },
    ];
    return <Table rowKey="id" columns={columns} dataSource={this.state.data} />;
  }
}

export default List;
