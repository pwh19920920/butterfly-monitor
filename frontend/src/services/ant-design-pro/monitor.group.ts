import { request } from '@umijs/max';

export async function monitorGroupQuery(params: API.PageParams & { name?: string }) {
  return request<API.Resp<API.MonitorGroup[]>>('/api/monitor/group', {
    method: 'GET',
    params,
  });
}

export async function monitorGroupQueryAll() {
  return request<API.Resp<API.MonitorGroup[]>>('/api/monitor/group/all', {
    method: 'GET',
  });
}

export async function monitorGroupCreate(data: API.MonitorGroup) {
  return request<API.Resp<string>>('/api/monitor/group', {
    method: 'POST',
    data,
  });
}

export async function monitorGroupUpdate(data: API.MonitorGroup) {
  return request<API.Resp<string>>('/api/monitor/group', {
    method: 'PUT',
    data,
  });
}
