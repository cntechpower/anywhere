let prefix = "";
// eslint-disable-next-line no-undef
if (process.env.NODE_ENV === "development") {
  prefix = "http://127.0.0.1:1114";
}

export default {
  proxyConfigListApi: prefix + "/api/v1/proxy/list",
  proxyConfigAddApi: prefix + "/api/v1/proxy/add",
  proxyConfigUpdateApi: prefix + "/api/v1/proxy/update",
  proxyAgentListApi: prefix + "/api/v1/agent/list",
  proxyAgentDelApi: prefix + "/api/v1/proxy/delete",
  userLoginApi: prefix + "/user_login",
  //support api
  getLocalIpAPI: prefix + "/api/v1/support/ip",
  getSummaryApi: prefix + "/api/v1/summary"
};
