import React from "react";
import "antd/dist/antd.css";
import "./App.css";
import { Layout, Menu } from "antd";
import { Link } from "react-router-dom";
import {
  CompassOutlined,
  PieChartOutlined,
  SettingOutlined,
} from "@ant-design/icons";

const { Header, Content, Footer, Sider } = Layout;
const { SubMenu } = Menu;

class SiderBar extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      collapsed: false,
    };
  }

  onCollapse = (collapsed) => {
    this.setState({ collapsed });
  };

  render() {
    const { collapsed } = this.state;
    const selectedKey = this.props.selectKey || "";
    const openKey = this.props.openKey || "";
    return (
      <Layout style={{ minHeight: "100vh" }}>
        <Sider collapsible collapsed={collapsed} onCollapse={this.onCollapse}>
          <div className="logo" onClick={this.goHome}>
            <h3 class="h3">Anywhere</h3>
          </div>
          <Menu
            theme="dark"
            defaultSelectedKeys={new Array(selectedKey)}
            defaultOpenKeys={new Array(openKey)}
            mode="inline"
          >
            <Menu.Item key="home" icon={<PieChartOutlined />}>
              <Link to="/home">系统状态</Link>
            </Menu.Item>

            <SubMenu key="configs" icon={<SettingOutlined />} title="配置管理">
              <Menu.Item key="add">
                <Link to="/configs/add">添加配置</Link>
              </Menu.Item>
              <Menu.Item key="list">
                <Link to="/configs/list">配置列表</Link>
              </Menu.Item>
            </SubMenu>
            <SubMenu key="stats" icon={<CompassOutlined />} title="状态管理">
              <Menu.Item key="conns">
                <Link to="/stats/conns">连接列表</Link>
              </Menu.Item>
            </SubMenu>
          </Menu>
        </Sider>
        <Layout className="site-layout">
          <Header className="site-layout-background" style={{ padding: 0 }} />
          <Content style={{ margin: "0 16px" }}>
            <div
              className="site-layout-background"
              style={{ padding: 24, minHeight: 360 }}
            >
              {this.props.children}
            </div>
          </Content>
          <Footer style={{ textAlign: "center" }}>
            ©2021 root@cntechpower.com
          </Footer>
        </Layout>
      </Layout>
    );
  }
}

export default SiderBar;
