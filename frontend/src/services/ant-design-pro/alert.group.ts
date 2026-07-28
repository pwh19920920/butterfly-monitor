import { request } from '@umijs/max';

export async function alertGroupQuery(params: API.PageParams & { name?: string }) {
  return request<API.Resp<API.AlertGroup[]>>('/api/alert/group', {
    method: 'GET',
    params,
  });
}

export async function alertGroupQueryAll() {
  return request<API.Resp<API.AlertGroup[]>>('/api/alert/group/all', {
    method: 'GET',
  });
}

export async function alertGroupCreate(data: { name: string; userIds: string[] }) {
  return request<API.Resp<string>>('/api/alert/group', {
    method: 'POST',
    data,
  });
}

export async function alertGroupUpdate(data: { id: string | number; name: string; userIds: string[] }) {
  return request<API.Resp<string>>('/api/alert/group', {
    method: 'PUT',
    data,
  });
}

export async function alertGroupUsers(id: string | number) {
  return request<API.Resp<string[]>>(`/api/alert/group/groupUser/${id}`, {
    method: 'GET',
  });
}
