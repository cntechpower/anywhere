import React from "react";
import ReactDOM from "react-dom";
import { BrowserRouter, Redirect, Route, Switch } from "react-router-dom";
import { ConfigProvider } from "antd";
import zhCN from "antd/es/locale/zh_CN";
import "antd/dist/antd.css";
import "./index.css";
import "./theme.css";
import "./antd-dark.css";
import App from "./App";
import Home from "./pages/home";
import ProxyConfigList from "./pages/configs/list";
import ProxyConfigAdd from "./pages/configs/add";
import StatsConnsList from "./pages/stats/connection";
import StatsAgentsList from "./pages/stats/agent";
import UserLogin from "./pages/user/login";
import NotFound from "./pages/NotFound";

ReactDOM.render(
  <React.StrictMode>
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <Switch>
          <Route path="/user/login" component={UserLogin} />
          <Route path="/">
            <App>
              <Switch>
                <Route exact path="/" render={() => <Redirect to="/home" />} />
                <Route path="/home" component={Home} />
                <Route path="/configs/add" component={ProxyConfigAdd} />
                <Route path="/configs/list" component={ProxyConfigList} />
                <Route path="/stats/conns" component={StatsConnsList} />
                <Route path="/stats/agents" component={StatsAgentsList} />
                <Route component={NotFound} />
              </Switch>
            </App>
          </Route>
        </Switch>
      </BrowserRouter>
    </ConfigProvider>
  </React.StrictMode>,
  document.getElementById("root")
);
