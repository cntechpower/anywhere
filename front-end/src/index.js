import React from "react";
import ReactDOM from "react-dom";
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";
import "./index.css";
import App from "./App";
import Home from "./pages/home";
import ProxyConfigList from "./pages/configs/list";
import ProxyConfigAdd from "./pages/configs/add";
import StatsConnsList from "./pages/stats/connection";
import UserLogin from "./pages/user/login";
import reportWebVitals from "./reportWebVitals";

ReactDOM.render(
  <React.StrictMode>
    <Router>
      {/* A <Switch> looks through its children <Route>s and
            renders the first one that matches the current URL. */}
      <Switch>
        <Route path="/home">
          <App openKey="home" selectKey="home">
            <Home />
          </App>
        </Route>

        <Route path="/configs/add">
          <App openKey="configs" selectKey="add">
            <ProxyConfigAdd />
          </App>
        </Route>

        <Route path="/configs/list">
          <App openKey="configs" selectKey="list">
            <ProxyConfigList />
          </App>
        </Route>

        <Route path="/stats/conns">
          <App openKey="stats" selectKey="conns">
            <StatsConnsList />
          </App>
        </Route>

        <Route path="/user/login">
          <UserLogin />
        </Route>
        <Route path="/">
          <App openKey="home" selectKey="home">
            <Home />
          </App>
        </Route>
      </Switch>
    </Router>
  </React.StrictMode>,
  document.getElementById("root")
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
