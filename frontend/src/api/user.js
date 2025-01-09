import request from '../utils/request'

export const login = (data) => {
  return request({
    url: '/login',
    method: 'post',
    data
  })
}

export const register = (data) => {
  return request({
    url: '/register',
    method: 'post',
    data
  })
}

export const getUserProfile = () => {
  return request({
    url: '/profile',
    method: 'get'
  })
}

export const updateBio = (data) => {
  return request({
    url: '/bio',
    method: 'post',
    data
  })
} 