import { createSelector } from 'reselect'
import { selector, deepEqualSelector } from '../common'

const profile = createSelector(selector, data => data.profile)
const isAdmin = createSelector(profile, data => data.is_admin)
const user = createSelector(profile, data => data.user)
const userName = createSelector(user, data => data.name)
const isEmailConfirmed = createSelector(user, data => data.is_confirmed)
const userEmail = createSelector(user, data => data.email)
const userId = createSelector(user, data => data.id)
const uploadLink = deepEqualSelector(profile, data => data.uploadLink)
const profileErrorMessage = deepEqualSelector(profile, data => data.profileErrorMessage)

export default {
    profile,
    isAdmin,
    isEmailConfirmed,
    userId,
    userName,
    userEmail,
    uploadLink,
    profileErrorMessage,
}