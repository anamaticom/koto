import { put } from 'redux-saga/effects'
import Actions from '@store/actions'
import { ApiTypes } from 'src/types'
import { API } from '@services/api'

export function* watchlogin(action: { type: string, payload: ApiTypes.Login }) {
  const response = yield API.authorization.login(action.payload)

  if (response.status === 200) {
    yield put(Actions.profile.getProfileRequest())
    yield put(Actions.authorization.loginSucces())
    localStorage.setItem('kotoIsLogged', 'true')
  } else {
    yield put(Actions.authorization.loginFailed(response?.error?.response?.data?.msg || 'Server error'))
  }
}

export function* watchlogout() {
  const response = yield API.authorization.logout()
  localStorage.clear()

  if (response.status === 200) {
    yield put(Actions.authorization.logoutSucces())
    window.location.reload()
  } else {
    localStorage.clear()
    window.location.reload()
  }
}

export function* watchGetAuthToken() {
  const response = yield API.authorization.getAuthToken()

  if (response.status === 200) {
    if (response.data?.token) {
      localStorage.setItem('kotoAuthToken', JSON.stringify(response.data?.token))
      localStorage.setItem('kotoAuthTokenDate', JSON.stringify(new Date()))
      yield put(Actions.authorization.getAuthTokenSucces(response.data?.token))
    }
  } else if (response.error.response.status === 401) {
    localStorage.clear()
    window.location.reload()
  }
}

export function* watchForgotPassword(action: { type: string, payload: ApiTypes.ForgotPassword }) {
  const response = yield API.authorization.forgotPassword(action.payload)

  if (response.status === 200) {
    yield put(Actions.authorization.forgotPasswordSuccess())
    yield put(Actions.common.setSuccessNotify('Sent successfully, check your email'))
  } else {
    yield put(Actions.authorization.forgotPasswordFailed(response?.error?.response?.data?.msg || 'Server error'))
  }
}

export function* watchResetPassword(action: { type: string, payload: ApiTypes.ResetPassword }) {
  const response = yield API.authorization.resetPassword(action.payload)

  if (response.status === 200) {
    yield put(Actions.authorization.resetPasswordSuccess())
    yield put(Actions.authorization.loginRequest({
      name: action.payload.name,
      password: action.payload.new_password,
    }))
  } else {
    yield put(Actions.authorization.resetPasswordFailed(response?.error?.response?.data?.msg || 'Server error'))
  }
}