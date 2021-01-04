/* eslint-disable react/prop-types */
import React from "react";
import { Link, withRouter } from "react-router-dom";
import "antd/dist/antd.css";
import "./GlobalLayout.css";
import { Layout, Menu, Icon } from "antd";
import { DOCUMENT_ROOT } from "../../constants/others";

const { Header, Content, Footer, Sider } = Layout;

const { SubMenu } = Menu;

class GlobalLayout extends React.Component {
  constructor(props) {
    super(props);

    this.goHome = this.goHome.bind(this);
    this.state.collapsed = false;
  }

  goHome() {
    this.props.history.push(DOCUMENT_ROOT);
  }
  state = {
    collapsed: false
  };

  onCollapse = collapsed => {
    console.log(collapsed);
    this.setState({ collapsed });
  };

  render() {
    const selectedKey = this.props.selectdKey || "";
    const openKey = this.props.openKey || "";
    return (
      <Layout style={{ minHeight: "100vh" }}>
        <Sider
          collapsible
          collapsed={this.state.collapsed}
          onCollapse={this.onCollapse}
        >
          <div className="logo" onClick={this.goHome} />
          <Menu
            theme="dark"
            defaultSelectedKeys={new Array(selectedKey)}
            defaultOpenKeys={new Array(openKey)}
            mode="inline"
          >
            <SubMenu
              key="note"
              title={
                <span>
                  <Icon type="unordered-list" />
                  <span>客户端配置</span>
                </span>
              }
            >
              <Menu.Item key="note_list">
                <Link to={DOCUMENT_ROOT + "/note/list"}>客户端列表</Link>
              </Menu.Item>
            </SubMenu>
            <SubMenu
              key="proxy"
              title={
                <span>
                  <Icon type="cloud" />
                  <span>穿透配置</span>
                </span>
              }
            >
              <Menu.Item key="proxy_add">
                <Link to={DOCUMENT_ROOT + "proxy/add"}>添加配置</Link>
              </Menu.Item>
              <Menu.Item key="proxy_list">
                <Link to={DOCUMENT_ROOT + "proxy/list"}>配置列表</Link>
              </Menu.Item>
            </SubMenu>
          </Menu>
        </Sider>
        <Layout>
          <Header style={{ background: "#fff", padding: 0 }} />
          <Content style={{ margin: "0 16px" }}>
            <div style={{ padding: 24, background: "#fff", minHeight: 360 }}>
              {this.props.children}
            </div>
          </Content>
          <Footer style={{ textAlign: "center" }}>
            Ant Design ©2020 Created by Ant UED
          </Footer>
        </Layout>
      </Layout>
    );
  }
}

export default withRouter(GlobalLayout);
