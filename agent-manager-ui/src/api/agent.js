import request from './request'

export function getAgents(params) {
  return request({
    url: '/agents',
    method: 'get',
    params
  })
}

export function getAgent(id) {
  return request({
    url: `/agents/${id}`,
    method: 'get'
  })
}

export function updateAgent(id, data) {
  return request({
    url: `/agents/${id}`,
    method: 'put',
    data
  })
}

export function deleteAgent(id) {
  return request({
    url: `/agents/${id}`,
    method: 'delete'
  })
}

export function getAgentStats() {
  return request({
    url: '/agents/stats',
    method: 'get'
  })
}
