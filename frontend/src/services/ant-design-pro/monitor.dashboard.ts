import { request } from '@umijs/max';

export async function monitorDashboardQuery(params: API.PageParams & { name?: string }) {
  return request<API.Resp<API.MonitorDashboard[]>>('/api/monitor/dashboard', {
    method: 'GET',
    params,
  });
}

export async function monitorDashboardQueryAll() {
  return request<API.Resp<API.MonitorDashboard[]>>('/api/monitor/dashboard/all', {
    method: 'GET',
  });
}

export async function monitorDashboardCreate(data: API.MonitorDashboard) {
  return request<API.Resp<string>>('/api/monitor/dashboard', {
    method: 'POST',
    data,
  });
}

export async function monitorDashboardUpdate(data: API.MonitorDashboard) {
  return request<API.Resp<string>>('/api/monitor/dashboard', {
    method: 'PUT',
    data,
  });
}

export async function monitorDashboardTask(id: string | undefined) {
  return request<API.Resp<API.MonitorDashboardTask[]>>(`/api/monitor/dashboard/task/${id}`, {
    method: 'GET',
  });
}

export async function monitorDashboardTaskSort(items: { id: string; sort: number }[]) {
  return request<API.Resp<string>>('/api/monitor/dashboard/taskSort', {
    method: 'PUT',
    data: { items },
  });
}
