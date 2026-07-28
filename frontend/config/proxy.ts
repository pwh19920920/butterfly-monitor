/**
 * 开发环境代理：/api -> 本地后端
 * 生产环境代理无效，由 nginx 处理（见 default.conf.template）
 */
export default {
  dev: {
    '/api/': {
      target: 'http://localhost:8088',
      changeOrigin: true,
    },
  },
  test: {
    '/api/': {
      target: 'http://localhost:8088',
      changeOrigin: true,
    },
  },
  pre: {
    '/api/': {
      target: 'your pre url',
      changeOrigin: true,
    },
  },
};
