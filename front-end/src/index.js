import React from "react";
import ReactDOM from "react-dom";
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";
import GlobalLayout from "./components/Layout/GlobalLayout";
import EmptyWrapper from "./components/Layout/EmptyWrapper";
import ProxyConfig from "./components/Anywhere/List";
import ProxyConfigAdd from "./components/Anywhere/Add";
import Summary from "./components/Summary/Summary";
import UserLogin from "./components/User/LoginApp";
import rootReducer from "./reducers";
import "./index.css";
import { createStore, applyMiddleware } from "redux";
import { Provider } from "react-redux";
import thunk from "redux-thunk";
import logger from "redux-logger";

import { DOCUMENT_ROOT } from "./constants/others";

const store = createStore(rootReducer, applyMiddleware(thunk, logger));

ReactDOM.render(
  <Provider store={store}>
    <Router>
      <Switch>
        <Route exact path={DOCUMENT_ROOT}>
          <GlobalLayout openKey="" selectKey="">
            <Summary />
          </GlobalLayout>
        </Route>
        <Route exact path={DOCUMENT_ROOT + "proxy/add"}>
          <GlobalLayout openKey="proxy" selectKey="proxy_add">
            <ProxyConfigAdd />
          </GlobalLayout>
        </Route>
        <Route exact path={DOCUMENT_ROOT + "proxy/list"}>
          <GlobalLayout openKey="proxy" selectKey="proxy_list">
            <ProxyConfig />
          </GlobalLayout>
        </Route>
        <Route exact path={DOCUMENT_ROOT + "user/login"}>
          <EmptyWrapper>
            <UserLogin />
          </EmptyWrapper>
        </Route>
      </Switch>
    </Router>
  </Provider>,
  document.getElementById("root")
);
