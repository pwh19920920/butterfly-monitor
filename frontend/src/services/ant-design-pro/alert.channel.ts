import { request } from '@umijs/max';

export async function alertChannelQuery(params: API.PageParams & { name?: string }) {
  return request<API.Resp<API.AlertChannel[]>>('/api/alert/channel', {
    method: 'GET',
    params,
  });
}

export async function alertChannelQueryAll() {
  return request<API.Resp<API.AlertChannel[]>>('/api/alert/channel/all', {
    method: 'GET',
  });
}

export async function alertChannelCreate(data: API.AlertChannel) {
  return request<API.Resp<string>>('/api/alert/channel', {
    method: 'POST',
    data,
  });
}

export async function alertChannelUpdate(data: API.AlertChannel) {
  return request<API.Resp<string>>('/api/alert/channel', {
    method: 'PUT',
    data,
  });
}

export async function alertChannelHandlers() {
  return request<API.Resp<API.AlertChannelHandler[]>>('/api/alert/channel/handlers', {
    method: 'GET',
  });
}
