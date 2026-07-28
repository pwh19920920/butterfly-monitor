import { request } from '@umijs/max';

/** 获取列表 GET /api/sys/role */
export async function sysRoleQuery(params: API.PageParams, options?: Record<string, any>) {
  return request<API.Resp<API.SysRole[]>>('/api/sys/role', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** 获取列表 GET /api/sys/role */
export async function sysRoleQueryAll() {
  return request<API.Resp<API.SysRole[]>>('/api/sys/role/all', {
    method: 'GET',
  });
}

/** POST /api/sys/role */
export async function sysRoleCreate(data: API.SysRole) {
  return request<API.Resp<API.SysRole[]>>('/api/sys/role', {
    method: 'POST',
    data: { ...data } || {},
  });
}

/** PUT /api/sys/role */
export async function sysRoleUpdate(data: API.SysRole) {
  return request<API.Resp<API.SysRole[]>>('/api/sys/role', {
    method: 'PUT',
    data: { ...data } || {},
  });
}

/** 删除 DELETE /api/sys/role/:id */
export async function sysRoleDelete(id: number | string) {
  return request<API.Resp<string>>(`/api/sys/role/${id}`, {
    method: 'DELETE',
  });
}

// 获取角色对应的权限串
export async function sysRolePermissionQuery(id: number | string) {
  return request<API.Resp<API.SysRolePermission[]>>(`/api/sys/role/permission/${id}`, {
    method: 'GET',
  });
}
