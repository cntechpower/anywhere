export interface ProxyConfig {
  id: number;
  user_name?: string;
  zone_name?: string;
  remote_port: number;
  local_addr?: string;
  is_whitelist_on?: boolean;
  whitelist_ips?: string;
  listen_type?: string;
  network_flow_local_to_remote_in_bytes?: number;
  network_flow_remote_to_local_in_bytes?: number;
  proxy_connect_count?: number;
  proxy_connect_reject_count?: number;
}

export interface AgentInfo {
  agentId?: string;
  userName?: string;
  zoneName?: string;
  agentAdminAddr?: string;
  lastAckSend?: string;
  lastAckRcv?: string;
  proxyConfigCount?: number;
}

export interface ZoneInfo {
  user_name?: string;
  zone_name?: string;
  agent_count?: number;
}

export interface ConnItem {
  id: number;
  src_remote_addr?: string;
  src_local_addr?: string;
  src_name?: string;
  dst_name?: string;
  dst_remote_addr?: string;
  dst_local_addr?: string;
}

export interface Summary {
  agent_total_count?: number;
  current_proxy_connection_count?: number;
  network_flow_total_count_in_bytes?: number;
  proxy_config_total_count?: number;
  proxy_connect_reject_count?: number;
  proxy_connect_total_count?: number;
  proxy_connect_reject_count_top10?: ProxyConfig[];
  proxy_network_flow_top10?: ProxyConfig[];
}

export interface LoginFormValues {
  username: string;
  password: string;
  otpcode?: string;
}

export interface WhiteListValues {
  white_list_enable?: boolean;
  white_list_ips?: string;
}

export interface ProxyFormValues extends WhiteListValues {
  zone: string;
  listen_type: "tcp" | "udp";
  remote_port: number;
  local_ip: string;
  local_port: number;
}

export interface CreatedProxyConfig {
  userName: string;
  zoneName: string;
  remotePort: number;
  localAddr: string;
  listenType: string;
  whiteList: boolean;
}

export interface BarItem {
  key?: string;
  label: string;
  value: number;
}
