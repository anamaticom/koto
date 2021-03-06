# User Hub API Reference

## Server Public Key

### Get the public key to check token signatures

```
POST http://central.koto.at/rpc.InfoService/PublicKey
Content-Type: application/json

{}
```

## Registration

### Register a new user and send email with confirmation link

```
POST https://central.koto.at/rpc.AuthService/Register
Content-Type: application/json

{
  "name":  "andrey",
  "email": "andrey@mail.com",
  "password":  "12345"
}

```

### Register a new invited user

```
POST https://central.koto.at/rpc.AuthService/Register
Content-Type: application/json

{
  "name":  "andrey",
  "email": "andrey@mail.com",
  "password":  "12345",
  "invite_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFuZHJleTQ0QG1haWwuY29tIiwiZXhwIjoxNjI4MTYwNTczLCJpZCI6ImJlYzc5ZmFiLWE4Y2UtNDdjOC1hODI2LTQ3ZjA4YTA2ZmY2ZiIsIm5hbWUiOiJhbmRyZXkxIiwic2NvcGUiOiJ1c2VyLWludml0ZSJ9.bBItLKq8fgw8yBEYXRNb8gt6DcDa_VchU5mR_fTdBkgcU5ZkMksZsm7PN2cEgBuW50SHKlaPjTfRHXfEjEH6JF3HOVmjs-sccjrv5VbQCxv6eFkAll6udBpznLWJ4VOxdOKrq19VXE41GWaMVApOMmbzwPoAIvrZxBq9CgwyfEY"
}
```

### Send email with confirmation link (explicitely)

```
POST https://central.koto.at/rpc.AuthService/SendConfirmLink
Content-Type: application/json
{}
```

### Confirm user by confirmation token

```
POST https://central.koto.at/rpc.AuthService/Confirm
Content-Type: application/json

{
  "token": "USER-CONFIRM-TOKEN"
}
```

## Authentication

### Login

```
POST https://central.koto.at/rpc.AuthService/Login
Content-Type: application/json

{
  "name": "andrey",
  "password": "12345",
  "remember_me": true
}
```

### Get current user info

```
POST https://central.koto.at/rpc.UserService/Me
Content-Type: application/json

{}
```

### Get a short-lived signed authentication token

```
POST https://central.koto.at/rpc.TokenService/Auth
Content-Type: application/json

{}
```

### Logout

```
POST https://central.koto.at/rpc.AuthService/Logout
Content-Type: application/json

{}
```

### Send email with "reset password" link

```
POST http://central.koto.at/rpc.AuthService/SendResetPasswordLink
Content-Type: application/json

{
  "name": "andrey1",
  "email": "andrey1@mail.com"
}
```

### Reset password

```
POST http://central.koto.at/rpc.AuthService/ResetPassword
Content-Type: application/json

{
  "reset_token": "RESET-TOKEN",
  "new_password": "54321"
}
```

## Message Hub Management

### Register a new message hub

```
POST https://central.koto.at/rpc.MessageHubService/Register
Content-Type: application/json

{
  "address": "https://localhost:12002",
  "details": "my cool hub",
  "post_limit": 2
}
```


### Get all hubs (admin access) or my hubs

```
POST https://central.koto.at/rpc.MessageHubService/Hubs
Content-Type: application/json

{}
```

### Verify a hub

```
POST https://central.koto.at/rpc.MessageHubService/Verify
Content-Type: application/json

{"hub_id":  "e60b3ff4-9ac0-4ba4-a45c-626c4eb29f75"}
```

### Approve a hub (admin access)

```
POST https://central.koto.at/rpc.MessageHubService/Approve
Content-Type: application/json

{"hub_id":  "e60b3ff4-9ac0-4ba4-a45c-626c4eb29f75"}
```

### Remove a hub

```
POST https://central.koto.at/rpc.MessageHubService/Remove
Content-Type: application/json

{"hub_id":  "ba7c6e53-dfea-46ec-b0ff-208f984393c4"}
```

### Set post limit for a hub

This will set limits for a hub as follows:
`"post_limit": 0` - no limit
`"post_limit": 1` - only admin can post
`"post_limit": 2` - only admin's friends can post
`"post_limit": 3` - admin's 2nd level of friends (friends of friends) can post
etc...

```
POST https://central.koto.at/rpc.MessageHubService/SetPostLimit
Content-Type: application/json

{
  "hub_id": "ebcaed9f-dbb4-40f6-982f-5f1fc5e3daf9",
  "post_limit": 2
}
```

## Invites

### Create invite


```
POST https://central.koto.at/rpc.InviteService/Create
Content-Type: application/json

{"friend": "andrey@mail.com"}
```

### Get invites from me

```
POST https://central.koto.at/rpc.InviteService/FromMe
Content-Type: application/json

{}
```

### Get invites for me

```
POST https://central.koto.at/rpc.InviteService/ForMe
Content-Type: application/json

{}
```

### Accept invite

```
POST https://central.koto.at//rpc.InviteService/Accept
Content-Type: application/json

{
  "inviter_id": "USER-ID"
}
```

### Reject invite

```
POST https://central.koto.at//rpc.InviteService/Reject
Content-Type: application/json

{
  "inviter_id": "USER-ID"
}
```

## Friends


### List of friends (for current user)

```
POST https://central.koto.at/rpc.UserService/Friends
Content-Type: application/json

{}
```

### List of friends of friends (for current user)

```
POST https://central.koto.at/rpc.UserService/FriendsOfFriends
Content-Type: application/json

{}
```

## Tokens

### Get a short-lived signed "post message" token

```
POST https://central.koto.at/rpc.TokenService/PostMessage
Content-Type: application/json

{}
```

### Get a short-lived signed "get messages" token

```
POST https://central.koto.at/rpc.TokenService/GetMessages
Content-Type: application/json

{}
```

## Blobs

### Get blob upload link

```
POST https://central.koto.at/rpc.BlobService/UploadLink
Content-Type: application/json

{
  "content_type": "image/png",
  "file_name": "image.png",
}
```

## Edit profile information for current user

```
POST https://central.koto.at/rpc.UserService/EditProfile
Content-Type: application/json

{
  "email_changed": true,
  "email": "andrey5@mmmail.com",
  "avatar_changed": true,
  "avatar_id": "BLOB-ID",
  "password_changed": true,
  "current_password": "12345",
  "new_password": "54321"
}
```

## Users


### Get users info

```
POST https://central.koto.at/rpc.UserService/Users
Content-Type: application/json

{
  "user_ids": ["bec79fab-a8ce-47c8-a826-47f08a06ff6f", "b002054c-2612-4906-b741-67f19e131126", "7b786381-00e6-4765-bc1f-8730c64d432d"]
}
```

### Get users info

```
POST http://central.koto.at/rpc.UserService/User
Content-Type: application/json

{
  "user_name": "andrey1"
}
```

## Notifications

### Notification counters (total, unread)

```
POST https://central.koto.at/rpc.NotificationService/Count
Content-Type: application/json

{}
```

### Notifications

```
POST https://central.koto.at/rpc.NotificationService/Notifications
Content-Type: application/json

{}
```

### Mark notifications as read

```
POST https://central.koto.at/rpc.NotificationService/MarkRead
Content-Type: application/json

{
  "last_known_id": "LAST-KNOWN-NOTIFICATION-ID"
}
```

### Clean notifications

```
POST https://central.koto.at/rpc.NotificationService/Clean
Content-Type: application/json

{
  "last_known_id": "LAST-KNOWN-NOTIFICATION-ID"
}
```

## FCM tokens

### Register a FCM token for current user

```
POST http://central.koto.at/rpc.UserService/RegisterFCMToken
Content-Type: application/json

{
  "token": "FCM-TOKEN",
  "device_id": "DEVICE-ID",
  "os": "android"
}
```
