import { request } from '@umijs/max';

export async function monitorTaskQuery(params: API.PageParams & Record<string, any>) {
  return request<API.Resp<API.MonitorTask[]>>('/api/monitor/task', {
    method: 'GET',
    params,
  });
}

// 查询任务详情，返回带 taskAlert/checkParams 的完整结构，供编辑态回显
export async function monitorTaskGetById(id: string | number) {
  return request<API.Resp<API.MonitorTaskQueryResponse>>(`/api/monitor/task/${id}`, {
    method: 'GET',
  });
}

export async function monitorTaskCreate(data: API.MonitorTaskCreate) {
  return request<API.Resp<string>>('/api/monitor/task', {
    method: 'POST',
    data,
  });
}

export async function monitorTaskUpdate(data: API.MonitorTaskCreate) {
  return request<API.Resp<string>>('/api/monitor/task', {
    method: 'PUT',
    data,
  });
}

export async function monitorTaskModifyTaskStatus(id: string | number, status: number) {
  return request<API.Resp<string>>(`/api/monitor/task/taskStatus/${id}/${status}`, {
    method: 'PUT',
  });
}

export async function monitorTaskModifyAlertStatus(id: string | number, status: number) {
  return request<API.Resp<string>>(`/api/monitor/task/alertStatus/${id}/${status}`, {
    method: 'PUT',
  });
}

export async function monitorTaskModifySampled(id: string | number, status: number) {
  return request<API.Resp<string>>(`/api/monitor/task/sampled/${id}/${status}`, {
    method: 'PUT',
  });
}
