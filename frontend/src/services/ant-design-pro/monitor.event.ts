import { request } from '@umijs/max';

export async function monitorTaskEventQuery(params: API.PageParams & Record<string, any>) {
  return request<API.Resp<API.MonitorTaskEvent[]>>('/api/monitor/task/event', {
    method: 'GET',
    params,
  });
}

export async function monitorTaskEventDeal(id: string | number, data?: { content?: string }) {
  return request<API.Resp<string>>(`/api/monitor/task/event/deal/${id}`, {
    method: 'POST',
    data: data || {},
  });
}

export async function monitorTaskEventComplete(id: string | number, data?: { content?: string }) {
  return request<API.Resp<string>>(`/api/monitor/task/event/complete/${id}`, {
    method: 'POST',
    data: data || {},
  });
}

export async function monitorTaskEventIgnore(id: string | number) {
  return request<API.Resp<string>>(`/api/monitor/task/event/ignore/${id}`, {
    method: 'POST',
    data: {},
  });
}

export async function monitorHomeCount() {
  return request<API.Resp<API.MonitorHomeCount>>('/api/monitor/homeCount', {
    method: 'GET',
  });
}
