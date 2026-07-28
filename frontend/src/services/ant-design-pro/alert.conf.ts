import { request } from '@umijs/max';

export async function alertConfQuery(params: API.PageParams & { confKey?: string }) {
  return request<API.Resp<API.AlertConf[]>>('/api/alert/conf', {
    method: 'GET',
    params,
  });
}

export async function alertConfCreate(data: API.AlertConf) {
  return request<API.Resp<string>>('/api/alert/conf', {
    method: 'POST',
    data,
  });
}

export async function alertConfUpdate(data: API.AlertConf) {
  return request<API.Resp<string>>('/api/alert/conf', {
    method: 'PUT',
    data,
  });
}
