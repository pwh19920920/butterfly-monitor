import { request } from '@umijs/max';

export async function monitorTaskQuery(params: API.PageParams & Record<string, any>) {
  return request<API.Resp<API.MonitorTask[]>>('/api/monitor/task', {
    method: 'GET',
    params,
  });
}

// 查询任务详情，返回带 taskAlert/checkParams 的完整结构，供编辑态回显
export async function monitorTaskGetById(id: string) {
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

export async function monitorTaskModifyTaskStatus(id: string, status: number) {
  return request<API.Resp<string>>(`/api/monitor/task/taskStatus/${id}/${status}`, {
    method: 'PUT',
  });
}

export async function monitorTaskModifyAlertStatus(id: string, status: number) {
  return request<API.Resp<string>>(`/api/monitor/task/alertStatus/${id}/${status}`, {
    method: 'PUT',
  });
}

export async function monitorTaskModifySampled(id: string, status: number) {
  return request<API.Resp<string>>(`/api/monitor/task/sampled/${id}/${status}`, {
    method: 'PUT',
  });
}

// 聚合预览：临时执行多行查询，返回结果列名，供前端勾选 label/value 维度（不落库）
export async function monitorTaskPreviewAggregate(data: Record<string, any>) {
  return request<API.Resp<API.MonitorTaskPreviewResponse>>('/api/monitor/task/previewAggregate', {
    method: 'POST',
    data,
  });
}
