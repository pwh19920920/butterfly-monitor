import { request } from '@umijs/max';

export async function monitorDatabaseQuery(params: API.PageParams & { name?: string; type?: number }) {
  return request<API.Resp<API.MonitorDatabase[]>>('/api/monitor/database', {
    method: 'GET',
    params,
  });
}

export async function monitorDatabaseQueryAll() {
  return request<API.Resp<API.MonitorDatabase[]>>('/api/monitor/database/all', {
    method: 'GET',
  });
}

export async function monitorDatabaseCreate(data: API.MonitorDatabase) {
  return request<API.Resp<string>>('/api/monitor/database', {
    method: 'POST',
    data,
  });
}

export async function monitorDatabaseUpdate(data: API.MonitorDatabase) {
  return request<API.Resp<string>>('/api/monitor/database', {
    method: 'PUT',
    data,
  });
}
