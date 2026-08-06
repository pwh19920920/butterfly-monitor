import { request } from '@umijs/max';

/** 获取列表 GET /api/sys/menu */
export async function sysMenuQuery() {
  return request<API.Resp<API.SysMenu[]>>('/api/sys/menu', {
    method: 'GET',
  });
}

/** 获取列表 GET /api/sys/menu */
export async function sysMenuQueryWithOption() {
  return request<API.Resp<API.SysMenu[]>>('/api/sys/menu/withOption', {
    method: 'GET',
  });
}

/** 获取列表 GET /api/sys/menu/:id/option */
export async function sysMenuOptionQuery(menuId: string) {
  return request<API.Resp<API.SysMenuOption[]>>(`/api/sys/menu/option/${menuId}`, {
    method: 'GET',
  });
}

/** 创建 POST /api/sys/menu */
export async function sysMenuCreate(data: API.SysMenu) {
  return request<API.Resp<string>>('/api/sys/menu', {
    method: 'POST',
    data: data ?? {},
  });
}

/** 删除 DELETE /api/sys/menu/:id */
export async function sysMenuDelete(id: string) {
  return request<API.Resp<string>>(`/api/sys/menu/${id}`, {
    method: 'DELETE',
  });
}

/** 更新 PUT /api/sys/menu */
export async function sysMenuUpdate(data: API.SysMenu) {
  return request<API.Resp<string>>('/api/sys/menu', {
    method: 'PUT',
    data: data ?? {},
  });
}
