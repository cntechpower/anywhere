import * as types from "../constants/ActionTypes";
import api from "../api/apiManager";
import axios from "axios";
import qs from "qs";

//Anywhere
export const resetProxyConfigProps = data => ({
  type: types.ANYWHERE_REFRESH_PROXY_CONFIG,
  data
});
export const getProxyConfig = () => ({
  type: types.ANYWHERE_GET_PROXY_CONFIG_LIST
});
export const delProxyConfig = id => ({
  type: types.ANYWHERE_DEL_PROXY_CONFIG,
  id
});

export const setError = error => ({
  type: types.ANYWHERE_SET_ERROR,
  error
});

export const fetchProxyConfigs = () => {
  return function(dispatch) {
    return fetch(api.proxyConfigListApi)
      .then(response => response.json())
      .then(
        response => {
          dispatch(resetProxyConfigProps(response));
        },
        error => {
          dispatch(setError(error));
        }
      );
  };
};

export const deleteProxyConfig = (
  agent_id,
  local_addr,
  fnOnSuccess,
  fnOnFail
) => {
  return function() {
    const options = {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      data: qs.stringify({
        agent_id: agent_id,
        local_addr: local_addr
      }),
      url: api.proxyAgentDelApi
    };
    return axios(options).then(
      () => {
        fnOnSuccess();
      },
      // Note: it's important to handle errors here
      // instead of a catch() block so that we don't swallow
      // exceptions from actual bugs in components.
      error => {
        fnOnFail();
      }
    );
  };
};
