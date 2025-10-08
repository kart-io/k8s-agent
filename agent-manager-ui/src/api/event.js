import request from './request'

export function getEvents(params) {
  return request({
    url: '/events',
    method: 'get',
    params
  })
}

export function getEvent(id) {
  return request({
    url: `/events/${id}`,
    method: 'get'
  })
}

export function getEventStats() {
  return request({
    url: '/events/stats',
    method: 'get'
  })
}
