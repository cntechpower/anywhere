/* eslint-disable import/no-anonymous-default-export */
let prefix = "";
// eslint-disable-next-line no-undef
if (process.env.NODE_ENV === "development") {
  prefix = "https://local.suya.host:1114";
}

export default {
  proxyConfigListApi: prefix + "/api/v1/proxy/list",
  proxyConfigAddApi: prefix + "/api/v1/proxy/add",
  proxyConfigUpdateApi: prefix + "/api/v1/proxy/update",
  proxyAgentListApi: prefix + "/api/v1/agent/list",
  proxyZoneListApi: prefix + "/api/v1/zone/list",
  proxyAgentDelApi: prefix + "/api/v1/proxy/delete",
  userLoginApi: prefix + "/user_login",
  //support api
  getLocalIpAPI: prefix + "/api/v1/support/ip",
  getSummaryApi: prefix + "/api/v1/summary",
};
