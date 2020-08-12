import React from 'react'
import AppBar from '@material-ui/core/AppBar'
import Toolbar from '@material-ui/core/Toolbar'
import { connect } from 'react-redux'
import Actions from '@store/actions'
import ExitToAppIcon from '@material-ui/icons/ExitToApp'
import { TopMenu } from './TopMenu'
import NotificationsActiveIcon from '@material-ui/icons/NotificationsActive'
import { history } from '@view/routes'
import selectors from '@selectors/index'
import { StoreTypes } from 'src/types'
import logo from './../../assets/images/logo-white.png'
import Avatar from '@material-ui/core/Avatar'
import { 
  TooltipStyle, 
  IconButtonStyled, 
  LogoWrapper, 
  TopBarRightSide,
  NotificationsCounter,
  NotificationsWrapper,
  AvatarWrapper,
  Logo,
} from './styles'

// @ts-ignore
const centralUrl: string = window.apiEndpoint

interface Props {
  notificationLength: number
  userId: string
  onLogout: () => void
}

const TopBar: React.SFC<Props> = React.memo((props) => {
  const { notificationLength, userId } = props

  const onLogoutClick = () => {
    history.push('/login')    
    props.onLogout()
  }

  return (
    <AppBar position="fixed" color="primary">
      <Toolbar>
        <LogoWrapper to="/messages">
          <Logo src={logo}/>
        </LogoWrapper>
        <TopBarRightSide>
          {Boolean(notificationLength) && <NotificationsWrapper to="/notifications">
            <NotificationsActiveIcon fontSize="small"/>
            <NotificationsCounter>{notificationLength} notifications</NotificationsCounter>
          </NotificationsWrapper>}
          <TopMenu />
          <AvatarWrapper to="/user-profile">
            <Avatar variant="rounded" src={`${centralUrl}/image/avatar/${userId}`}  />
          </AvatarWrapper>
          <TooltipStyle title={`Logout`}>
            <IconButtonStyled onClick={onLogoutClick}>
              <ExitToAppIcon fontSize="small" />
            </IconButtonStyled>
          </TooltipStyle>
        </TopBarRightSide>
      </Toolbar>
    </AppBar>
  )
})

type StateProps = Pick<Props, 'notificationLength' | 'userId'>
const mapStateToProps = (state: StoreTypes): StateProps => ({
  notificationLength: selectors.notifications.notificationLength(state),
  userId: selectors.profile.userId(state),
})

type DispatchProps = Pick<Props, 'onLogout'>
const mapDispatchToProps = (dispatch): DispatchProps => ({
  onLogout: () => dispatch(Actions.authorization.logoutRequest()),
})

export default connect(mapStateToProps, mapDispatchToProps)(TopBar)
