import * as types from "../constants/ActionTypes";
import axios from "axios";
import qs from "qs";
import api from "../api/apiManager";

//Anywhere
export const startAddProxyConfig = () => ({
  type: types.ANYWHERE_START_ADD_PROXY_CONFIG
});
export const addProxyConfig = (id, config) => ({
  type: types.ANYWHERE_ADD_PROXY_CONFIG,
  id,
  config
});
export const addProxyConfigSuccess = config => ({
  type: types.ANYWHERE_ADD_PROXY_CONFIG_SUCCESS,
  config
});
export const addProxyConfigFail = (error, config) => ({
  type: types.ANYWHERE_ADD_PROXY_CONFIG_FAIL,
  config,
  error
});
export const clearErrorState = () => ({
  type: types.ANYWHERE_CLEAR_ERROR_STATE
});

export const getAgentsList = () => ({
  type: types.ANYWHERE_ADD_GET_AGENTS_LIST
});

export const resetAgentListProps = data => ({
  type: types.ANYWHERE_ADD_REFRESH_AGENTS_LIST,
  data
});

export const refreshLocalIp = ip => ({
  type: types.ANYWHERE_REFRESH_GET_LOCAL_IP,
  ip
});

export const getLocalIp = () => ({
  type: types.ANYWHERE_UPDATE_GET_LOCAL_IP
});

export const postProxyConfig = config => {
  return function(dispatch) {
    const options = {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      data: qs.stringify({
        agent_id: config.agent_id,
        remote_port: config.remote_port,
        local_addr: config.local_ip + ":" + config.local_port,
        white_list_enable: config.white_list_enable,
        white_list_ips: config.white_list_ips || ""
      }),
      url: api.proxyConfigAddApi
    };
    return axios(options).then(
      () => {
        dispatch(addProxyConfigSuccess(config));
      },
      // Note: it's important to handle errors here
      // instead of a catch() block so that we don't swallow
      // exceptions from actual bugs in components.
      error => {
        dispatch(addProxyConfigFail(error, config));
      }
    );
  };
};

export const fetchAgentLists = () => {
  return function(dispatch) {
    return fetch(api.proxyAgentListApi)
      .then(response => response.json())
      .then(
        response => {
          dispatch(resetAgentListProps(response));
        },
        error => {
          dispatch(addProxyConfigFail(error));
        }
      );
  };
};

export const fetchGetLocalIp = fnOnSuccess => {
  return function(dispatch) {
    return fetch(api.getLocalIpAPI)
      .then(response => response.json())
      .then(
        response => {
          fnOnSuccess(response);
          // dispatch(refreshLocalIp(response));
        },
        error => {
          dispatch(addProxyConfigFail(error));
        }
      );
  };
};
