import {
  ANYWHERE_START_ADD_PROXY_CONFIG,
  ANYWHERE_ADD_PROXY_CONFIG_SUCCESS,
  ANYWHERE_ADD_PROXY_CONFIG_FAIL,
  ANYWHERE_CLEAR_ERROR_STATE,
  ANYWHERE_ADD_REFRESH_AGENTS_LIST,
  ANYWHERE_ADD_GET_AGENTS_LIST,
  ANYWHERE_REFRESH_GET_LOCAL_IP
} from "../constants/ActionTypes";

const initialState = {
  init: true,
  creating: false,
  createdOk: false,
  error: null,
  config: null,
  agentsLoading: true,
  localIp: "",
  agents: []
};

export default function anywhereList(state = initialState, action) {
  switch (action.type) {
    case ANYWHERE_START_ADD_PROXY_CONFIG:
      return Object.assign({}, state, {
        init: false,
        creating: true
      });
    case ANYWHERE_ADD_PROXY_CONFIG_FAIL:
      return Object.assign({}, state, {
        init: false,
        creating: false,
        createdOk: false,
        error: action.error,
        config: action.config
      });
    case ANYWHERE_ADD_PROXY_CONFIG_SUCCESS:
      return Object.assign({}, state, {
        init: false,
        creating: false,
        createdOk: true,
        config: action.config
      });
    case ANYWHERE_ADD_GET_AGENTS_LIST:
      return Object.assign({}, state, {
        agentsLoading: true
      });
    case ANYWHERE_ADD_REFRESH_AGENTS_LIST:
      return Object.assign({}, state, {
        agentsLoading: false,
        agents: action.data
      });
    case ANYWHERE_CLEAR_ERROR_STATE:
      return Object.assign({}, state, {
        init: true,
        error: null
      });
    case ANYWHERE_REFRESH_GET_LOCAL_IP:
      return Object.assign({}, state, {
        localIp: action.ip
      });
    default:
      return state;
  }
}
