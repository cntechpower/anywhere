const prefix = import.meta.env.VITE_API_PREFIX ?? "";

export const apis = {
  proxyConfigListApi: `${prefix}/api/v1/proxy/list`,
  proxyConfigAddApi: `${prefix}/api/v1/proxy/add`,
  proxyConfigUpdateApi: `${prefix}/api/v1/proxy/update`,
  proxyAgentListApi: `${prefix}/api/v1/agent/list`,
  proxyZoneListApi: `${prefix}/api/v1/zone/list`,
  proxyAgentDelApi: `${prefix}/api/v1/proxy/delete`,
  userLoginApi: `${prefix}/user_login`,
  userLogoutApi: `${prefix}/user_logout`,
  // support api
  getLocalIpAPI: `${prefix}/api/v1/support/ip`,
  getSummaryApi: `${prefix}/api/v1/summary`,
  // stats api
  statsGetConnectionListApi: `${prefix}/api/v1/connection/list`,
  statsKillConnectionApi: `${prefix}/api/v1/connection/kill`,
};

export default apis;
