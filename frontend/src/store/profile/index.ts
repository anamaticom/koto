import { Types } from './actions'
import { ApiTypes } from './../../types'

const profile = localStorage.getItem('kotoProfile')
const user = profile ? JSON.parse(profile)?.user : {
  id: '',
  name: '',
}

export interface State extends ApiTypes.Profile {}

const initialState: State = {
  user: user,
}

const reducer = (state = initialState, action) => {
  switch (action.type) {
    case Types.GET_PROFILE_SUCCESS: {
      return {
        ...state, ...action.payload
      }
    }
    default: return state
  }
}

export default reducer