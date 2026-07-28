import type { RequestOptions } from '@@/plugin-request/request';
import type { RequestConfig } from '@umijs/max';
import { history } from '@umijs/max';
import { notification } from 'antd';

const loginPath = '/user/login';

const codeMessage: Record<number, string> = {
  200: '服务器成功返回请求的数据。',
  201: '新建或修改数据成功。',
  202: '一个请求已经进入后台排队（异步任务）。',
  204: '删除数据成功。',
  400: '发出的请求有错误，服务器没有进行新建或修改数据的操作。',
  401: '用户没有权限（令牌、用户名、密码错误）。',
  403: '用户得到授权，但是访问是被禁止的。',
  404: '发出的请求针对的是不存在的记录，服务器没有进行操作。',
  405: '请求方法不被允许。',
  406: '请求的格式不可得。',
  410: '请求的资源被永久删除，且不会再得到的。',
  422: '当创建一个对象时，发生一个验证错误。',
  500: '服务器发生错误，请检查服务器。',
  502: '网关错误。',
  503: '服务不可用，服务器暂时过载或维护。',
  504: '网关超时。',
};

/**
 * butterfly 业务请求配置：Bearer token + 401 回登录
 * 不使用模板里的 success/errorCode 业务码约定（后端为 status/message/data）
 */
export const errorConfig: RequestConfig = {
  errorConfig: {
    errorHandler: (error: any, opts: any) => {
      if (opts?.skipErrorHandler) throw error;

      const response = error?.response;
      const data = error?.data ?? response?.data;

      if (response?.status) {
        if (history.location.pathname === loginPath) {
          throw error;
        }

        let errorText =
          codeMessage[response.status as number] || response.statusText;
        if (data?.message) {
          errorText = data.message;
        }

        notification.error({
          message: `请求错误: ${response.status}`,
          description: errorText,
        });

        if (response.status === 401) {
          const search = history.location.search || '';
          const params = new URLSearchParams(search);
          const existingRedirect = params.get('redirect');
          if (existingRedirect) {
            window.location.href = `${loginPath}?redirect=${encodeURIComponent(existingRedirect)}`;
          } else {
            window.location.href = `${loginPath}?redirect=${encodeURIComponent(
              history.location.pathname + history.location.search,
            )}`;
          }
          localStorage.clear();
        }
      } else if (data?.message) {
        notification.error({
          message: '请求错误',
          description: data.message,
        });
      } else if (typeof navigator !== 'undefined' && !navigator.onLine) {
        notification.error({
          message: '网络异常',
          description: '您的网络发生异常，无法连接服务器',
        });
      } else {
        notification.error({
          message: '网络异常',
          description: '您的网络发生异常，无法连接服务器',
        });
      }
      throw error;
    },
  },

  requestInterceptors: [
    (config: RequestOptions) => {
      const headers = {
        ...config.headers,
      } as Record<string, string>;
      const token = localStorage.getItem('token');
      if (token) {
        headers.Authorization = token;
      }
      if (!headers['Content-Type'] && !headers['content-type']) {
        headers['Content-Type'] = 'application/json';
      }
      return { ...config, headers };
    },
  ],

  responseInterceptors: [],
};
