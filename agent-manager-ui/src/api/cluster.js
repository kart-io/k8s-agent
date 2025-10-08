import request from './request'

export function getClusters(params) {
  return request({
    url: '/clusters',
    method: 'get',
    params
  })
}

export function getCluster(id) {
  return request({
    url: `/clusters/${id}`,
    method: 'get'
  })
}

export function createCluster(data) {
  return request({
    url: '/clusters',
    method: 'post',
    data
  })
}

export function updateCluster(id, data) {
  return request({
    url: `/clusters/${id}`,
    method: 'put',
    data
  })
}

export function deleteCluster(id) {
  return request({
    url: `/clusters/${id}`,
    method: 'delete'
  })
}
