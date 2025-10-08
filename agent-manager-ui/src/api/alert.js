import request from './request'

export function getAlerts(params) {
  return request({
    url: '/alerts',
    method: 'get',
    params
  })
}

export function getAlert(id) {
  return request({
    url: `/alerts/${id}`,
    method: 'get'
  })
}

export function createAlert(data) {
  return request({
    url: '/alerts',
    method: 'post',
    data
  })
}

export function updateAlert(id, data) {
  return request({
    url: `/alerts/${id}`,
    method: 'put',
    data
  })
}

export function deleteAlert(id) {
  return request({
    url: `/alerts/${id}`,
    method: 'delete'
  })
}
