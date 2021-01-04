import * as types from "../constants/ActionTypes";
import api from "../api/apiManager";
import qs from "qs";
import axios from "axios";

export const userLoginStart = () => ({
  type: types.USER_LOGIN_START
});

export const userLoginSuccess = () => ({
  type: types.USER_LOGIN_SUCCESS
});

export const userLoginFailed = error => ({
  type: types.USER_LOGIN_FAILED,
  error
});

export const userLogout = () => ({
  type: types.USER_LOGOUT
});

export const userClearError = () => ({
  type: types.USER_LOGIN_CLEAR_ERROR
});

export const userLogin = (username, password, otpcode) => {
  return function(dispatch) {
    dispatch(userLoginStart());
    const options = {
      method: "POST",
      headers: { "content-type": "application/x-www-form-urlencoded" },
      data: qs.stringify({
        username: username,
        password: password,
        otpcode: otpcode
      }),
      url: api.userLoginApi
    };
    return axios(options).then(
      () => {
        dispatch(userLoginSuccess());
      },
      error => {
        dispatch(userLoginFailed(error));
      }
    );
  };
};
