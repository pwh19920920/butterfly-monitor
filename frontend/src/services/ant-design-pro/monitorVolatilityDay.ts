import { request } from '@umijs/max';

export async function volatilityDayQueryAll() {
  return request<API.Resp<API.MonitorVolatilityDay[]>>('/api/monitor/volatilityDay', {
    method: 'GET',
  });
}

export async function volatilityDayBatchCreate(data: API.MonitorVolatilityDayBatchCreateRequest) {
  return request<API.Resp<string>>('/api/monitor/volatilityDay/batch', {
    method: 'POST',
    data,
  });
}

export async function volatilityDayUpdate(id: string | undefined, data: API.MonitorVolatilityDay) {
  return request<API.Resp<string>>(`/api/monitor/volatilityDay/${id}`, {
    method: 'PUT',
    data,
  });
}

export async function volatilityDayDelete(id: string | undefined) {
  return request<API.Resp<string>>(`/api/monitor/volatilityDay/${id}`, {
    method: 'DELETE',
  });
}
