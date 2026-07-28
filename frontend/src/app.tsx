import {
  AlertOutlined,
  CrownOutlined,
  DashboardOutlined,
  SmileOutlined,
  TableOutlined,
} from '@ant-design/icons';
import type { Settings as LayoutSettings } from '@ant-design/pro-components';
import type { RequestConfig, RunTimeLayoutConfig } from '@umijs/max';
import { history, Link } from '@umijs/max';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import React from 'react';

dayjs.extend(relativeTime);

import { AvatarDropdown, ErrorBoundary, Footer } from '@/components';
import { currentUser as queryCurrentUser } from '@/services/ant-design-pro/login';
import defaultSettings from '../config/defaultSettings';
import { errorConfig } from './requestErrorConfig';

const loginPath = '/user/login';

const iconMap: Record<string, React.ReactNode> = {
  smile: <SmileOutlined />,
  crown: <CrownOutlined />,
  table: <TableOutlined />,
  dashboard: <DashboardOutlined />,
  alert: <AlertOutlined />,
};

/**
 * @see https://umijs.org/docs/api/runtime-config#getinitialstate
 */
export async function getInitialState(): Promise<{
  settings?: Partial<LayoutSettings>;
  currentUser?: API.SysUser;
  loading?: boolean;
  fetchUserInfo?: () => Promise<API.SysUser | undefined>;
}> {
  const fetchUserInfo = async () => {
    try {
      const msg = await queryCurrentUser({
        skipErrorHandler: true,
      });
      return msg.data;
    } catch (_error) {
      const { pathname, search, hash } = history.location;
      if (!pathname.includes(loginPath)) {
        history.replace(
          `${loginPath}?redirect=${encodeURIComponent(pathname + search + hash)}`,
        );
      }
    }
    return undefined;
  };

  const { location } = history;
  // 登录页且无 token 时不拉用户；有 token 时仍尝试拉取
  if (!location.pathname.includes(loginPath) || localStorage.getItem('token')) {
    const currentUser = await fetchUserInfo();
    return {
      fetchUserInfo,
      currentUser,
      settings: defaultSettings as Partial<LayoutSettings>,
    };
  }

  return {
    fetchUserInfo,
    settings: defaultSettings as Partial<LayoutSettings>,
  };
}

// ProLayout 支持的 api https://procomponents.ant.design/components/layout
export const layout: RunTimeLayoutConfig = ({ initialState }) => {
  return {
    menuItemRender: (item, dom) => {
      if (item.path) {
        return (
          <Link to={item.path} prefetch>
            {dom}
          </Link>
        );
      }
      return dom;
    },
    actionsRender: () => [],
    avatarProps: {
      src: initialState?.currentUser?.avatar,
      title: initialState?.currentUser?.name,
      render: (_, avatarChildren) => (
        <AvatarDropdown>{avatarChildren}</AvatarDropdown>
      ),
    },
    waterMarkProps: {
      content: initialState?.currentUser?.name,
    },
    footerRender: () => <Footer />,
    onPageChange: () => {
      const { location } = history;
      if (
        !initialState?.currentUser &&
        !location.pathname.includes(loginPath)
      ) {
        history.replace(
          `${loginPath}?redirect=${encodeURIComponent(
            location.pathname + location.search + location.hash,
          )}`,
        );
        return;
      }

      if (initialState?.currentUser) {
        const params = new URLSearchParams(location.search || '');
        const redirect = params.get('redirect');
        if (redirect) {
          if (redirect.includes('://')) {
            window.location.href = redirect;
            return;
          }
          history.push(redirect || '/');
        }
      }
    },
    links: [],
    ErrorBoundary,
    menuHeaderRender: undefined,
    menuDataRender: (menuData) =>
      menuData.map((item) => {
        if (typeof item.icon === 'string') {
          return {
            ...item,
            icon: iconMap[item.icon] ?? item.icon,
          };
        }
        return item;
      }),
    menu: {
      params: {
        userId: initialState?.currentUser?.id,
      },
      request: async (_params, defaultMenuData) => {
        return initialState?.currentUser?.menus || defaultMenuData;
      },
    },
    ...initialState?.settings,
  };
};

export const request: RequestConfig = {
  // 开发走 proxy，生产由 nginx 反代同源 /api
  baseURL: '',
  ...errorConfig,
};

export function rootContainer(container: React.ReactNode) {
  return <ErrorBoundary>{container}</ErrorBoundary>;
}
