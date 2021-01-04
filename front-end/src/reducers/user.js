import {
  USER_LOGIN_FAILED,
  USER_LOGIN_SUCCESS,
  USER_LOGIN_START,
  USER_LOGIN_CLEAR_ERROR
} from "../constants/ActionTypes";

const initialState = {
  init: true,
  loading: false,
  isLogin: false,
  isLoginError: false,
  error: null
};
export default function user(state = initialState, action) {
  switch (action.type) {
    case USER_LOGIN_SUCCESS:
      return Object.assign({}, state, {
        init: false,
        loading: false,
        isLogin: true
      });
    case USER_LOGIN_START:
      return Object.assign({}, state, {
        init: false,
        loading: true
      });
    case USER_LOGIN_FAILED:
      return Object.assign({}, state, {
        init: false,
        loading: false,
        isLogin: false,
        isLoginError: true,
        error: action.error
      });
    case USER_LOGIN_CLEAR_ERROR:
      return initialState;
    default:
      return state;
  }
}
