import request from './request'

export function getCommands(params) {
  return request({
    url: '/commands',
    method: 'get',
    params
  })
}

export function getCommand(id) {
  return request({
    url: `/commands/${id}`,
    method: 'get'
  })
}

export function createCommand(data) {
  return request({
    url: '/commands',
    method: 'post',
    data
  })
}

export function executeCommand(id) {
  return request({
    url: `/commands/${id}/execute`,
    method: 'post'
  })
}
