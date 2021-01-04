import {
  ANYWHERE_REFRESH_PROXY_CONFIG,
  ANYWHERE_DEL_PROXY_CONFIG,
  ANYWHERE_SET_ERROR
} from "../constants/ActionTypes";

const initialState = {
  data: [""],
  loading: true,
  error: null
};

export default function anywhere(state = initialState, action) {
  switch (action.type) {
    case ANYWHERE_DEL_PROXY_CONFIG:
      return Object.assign({}, state, {
        loading: false,
        error: "REDUCER: DEL_PROXY_CONFIG"
      });
    case ANYWHERE_REFRESH_PROXY_CONFIG:
      return Object.assign({}, state, {
        loading: false,
        error: null,
        data: action.data
      });
    case ANYWHERE_SET_ERROR:
      return Object.assign({}, state, {
        loading: false,
        error: action.error.message
      });
    default:
      return state;
  }
}
