import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import { NavLink, useLocation } from "react-router-dom";
import { Dropdown } from "antd";
import type { MenuProps } from "antd";
import {
  ApiOutlined,
  ClusterOutlined,
  DashboardOutlined,
  LinkOutlined,
  LogoutOutlined,
  MenuOutlined,
  PlusCircleOutlined,
  UserOutlined,
} from "@ant-design/icons";
import Brand from "./components/Brand";
import apis from "./apis/apis";
import { apiPostForm } from "./api/client";

interface NavItem {
  to: string;
  label: string;
  icon: ReactNode;
  exact?: boolean;
}

interface NavSection {
  label: string;
  items: NavItem[];
}

const NAV_SECTIONS: NavSection[] = [
  {
    label: "总览",
    items: [
      {
        to: "/home",
        label: "概览",
        icon: <DashboardOutlined />,
        exact: true,
      },
    ],
  },
  {
    label: "代理",
    items: [
      { to: "/configs/list", label: "配置列表", icon: <ApiOutlined /> },
      { to: "/configs/add", label: "新增配置", icon: <PlusCircleOutlined /> },
    ],
  },
  {
    label: "观测",
    items: [
      { to: "/stats/agents", label: "节点", icon: <ClusterOutlined /> },
      { to: "/stats/conns", label: "连接", icon: <LinkOutlined /> },
    ],
  },
];

const PAGE_META: Record<string, { title: string; subtitle: string }> = {
  "/home": { title: "概览", subtitle: "系统运行状态、流量与拒绝连接概览" },
  "/configs/list": {
    title: "代理配置",
    subtitle: "管理对外端口与内网目标的映射关系",
  },
  "/configs/add": {
    title: "新增配置",
    subtitle: "为可用区创建一条 TCP / UDP 隧道",
  },
  "/stats/agents": {
    title: "节点",
    subtitle: "Agent 心跳、可用区分布与健康状态",
  },
  "/stats/conns": { title: "连接", subtitle: "当前活跃的隧道与客户端会话" },
};

function currentUser(): string {
  try {
    return window.localStorage.getItem("anywhere.user") || "admin";
  } catch {
    return "admin";
  }
}

interface AppProps {
  children: ReactNode;
}

export default function App({ children }: AppProps) {
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const meta = PAGE_META[location.pathname] ?? {
    title: "控制台",
    subtitle: "",
  };
  const user = currentUser();
  const initials = user.slice(0, 2).toUpperCase();

  useEffect(() => {
    setOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    document.title = `${meta.title} · Anywhere 隧道控制台`;
  }, [meta.title]);

  useEffect(() => {
    if (!open) {
      return undefined;
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setOpen(false);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open]);

  const handleLogout = async () => {
    try {
      await apiPostForm(apis.userLogoutApi, {});
    } finally {
      try {
        window.localStorage.removeItem("anywhere.user");
      } catch {
        /* storage may be unavailable */
      }
      window.location.assign("/user/login");
    }
  };

  const userMenu: MenuProps = {
    items: [
      {
        key: "user",
        icon: <UserOutlined />,
        label: user,
        disabled: true,
      },
      { type: "divider" },
      {
        key: "logout",
        icon: <LogoutOutlined />,
        label: "退出登录",
        onClick: handleLogout,
      },
    ],
  };

  return (
    <div className="shell">
      <a className="skip-link" href="#main-content">
        跳到主要内容
      </a>

      {open && (
        <div
          className="sidebar-scrim"
          role="presentation"
          onClick={() => setOpen(false)}
        />
      )}

      <aside className={`sidebar${open ? " open" : ""}`} aria-label="主导航">
        <div className="sidebar-brand">
          <Brand />
        </div>

        <nav className="sidebar-nav">
          {NAV_SECTIONS.map((section) => (
            <div key={section.label}>
              <div className="nav-section">{section.label}</div>
              {section.items.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  exact={item.exact}
                  className="nav-link"
                  activeClassName="active"
                >
                  {item.icon}
                  <span>{item.label}</span>
                </NavLink>
              ))}
            </div>
          ))}
        </nav>

        <div className="sidebar-foot">
          <div className="live-line">
            <span className="pulse-dot" aria-hidden="true" />
            控制台在线
          </div>
          <div>安全隧道 · mTLS</div>
        </div>
      </aside>

      <div className="main">
        <header className="topbar">
          <button
            type="button"
            className="icon-btn menu-btn"
            aria-label={open ? "关闭导航" : "打开导航"}
            aria-expanded={open}
            onClick={() => setOpen((value) => !value)}
          >
            <MenuOutlined />
          </button>

          <div className="topbar-title">
            <nav className="crumbs" aria-label="面包屑">
              <span>控制台</span>
              <em>/</em>
              <strong>{meta.title}</strong>
            </nav>
            <span>{meta.subtitle}</span>
          </div>

          <div className="topbar-actions">
            <span className="live-chip">
              <span className="pulse-dot" aria-hidden="true" />
              实时数据
            </span>
            <Dropdown
              menu={userMenu}
              trigger={["click"]}
              placement="bottomRight"
            >
              <button type="button" className="user-chip" aria-label="账户菜单">
                <span className="avatar" aria-hidden="true">
                  {initials}
                </span>
                <span>{user}</span>
              </button>
            </Dropdown>
          </div>
        </header>

        <main className="content" id="main-content">
          {children}
        </main>
      </div>
    </div>
  );
}
