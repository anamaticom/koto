import React, { ChangeEvent } from 'react'
import ListItem from '@material-ui/core/ListItem'
import Paper from '@material-ui/core/Paper'
import Divider from '@material-ui/core/Divider'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemAvatar from '@material-ui/core/ListItemAvatar'
import IconButton from '@material-ui/core/IconButton'
import Tooltip from '@material-ui/core/Tooltip'
import CheckCircleIcon from '@material-ui/icons/CheckCircle'
import CancelIcon from '@material-ui/icons/Cancel'
import { connect } from 'react-redux'
import Actions from '@store/actions'
import { StoreTypes, ApiTypes } from 'src/types'
import selectors from '@selectors/index'
import { getAvatarUrl } from '@services/avatarUrl'
import {
  UsersWrapper,
  ListStyled,
  SearchWrapper,
  EmptyMessage,
  UserName,
  IconButtonGreen,
  PageWrapper,
  AvatarStyled,
  SearchInput,
  SearchIconStyled,
} from './styles'

export interface Props {
  invitations: ApiTypes.Friends.Invitation[]
  onGetInvitations: () => void
  onAcceptInvitation: (data: ApiTypes.Friends.InvitationAccept) => void
  onRejectInvitation: (data: ApiTypes.Friends.InvitationReject) => void
}

interface State {
  pendingInvitations: ApiTypes.Friends.Invitation[]
  searchResult: ApiTypes.Friends.Invitation[]
  searchValue: string
}

export class Invitations extends React.Component<Props, State> {

  state = {
    searchResult: [],
    searchValue: '',
    pendingInvitations: [],
  }

  searchInputRef = React.createRef<HTMLInputElement>()

  static getDerivedStateFromProps(newProps: Props) {
    return {
      pendingInvitations: newProps.invitations?.length && newProps.invitations.filter(
        item => !item.accepted_at && !item.rejected_at
      )
    }
  }

  showEmptyListMessage = () => {
    const { searchValue } = this.state

    if (searchValue) {
      return <EmptyMessage>No one's been found.</EmptyMessage>
    } else {
      return <EmptyMessage>You don't have any invitations yet.</EmptyMessage>
    }
  }

  mapInvitations = (invitations: ApiTypes.Friends.Invitation[]) => {
    const { onAcceptInvitation, onRejectInvitation } = this.props

    if (!invitations || !invitations?.length) {
      return this.showEmptyListMessage()
    }

    return invitations.map(item => {
      const { friend_id, friend_name } = item
      return (
        <div key={friend_id}>
          <ListItem alignItems="center">
            <ListItemAvatar>
              <AvatarStyled alt={friend_name} src={getAvatarUrl(friend_id)} />
            </ListItemAvatar>
            <ListItemText primary={<UserName>{friend_name}</UserName>} />
            <Tooltip title={`Accept the invitation`}>
              <IconButtonGreen onClick={() => onAcceptInvitation({ inviter_id: friend_id })}>
                <CheckCircleIcon />
              </IconButtonGreen>
            </Tooltip>
            <Tooltip title={`Decline the invitation`}>
              <IconButton color="secondary" onClick={() => onRejectInvitation({ inviter_id: friend_id })}>
                <CancelIcon />
              </IconButton>
            </Tooltip>
          </ListItem>
          <Divider variant="inset" component="li" />
        </div>
      )
    })
  }

  onSearch = (event: ChangeEvent<HTMLInputElement>) => {
    const { pendingInvitations } = this.state
    const { value } = event.currentTarget

    this.setState({
      searchValue: value,
      searchResult: pendingInvitations.filter(
        (item: ApiTypes.Friends.Invitation) => {
          return item.friend_name.toLowerCase().includes(value.toLowerCase())
        }
      )
    })
  }

  componentDidMount() {
    this.props.onGetInvitations()
  }

  render() {
    const { pendingInvitations, searchResult, searchValue } = this.state

    return (
      <PageWrapper>
        <UsersWrapper>
          <Paper>
            <SearchWrapper>
              <SearchIconStyled onClick={() => this.searchInputRef?.current?.focus()} />
              <SearchInput
                ref={this.searchInputRef}
                id="filter"
                placeholder="Filter"
                onChange={this.onSearch}
                value={searchValue}
              />
            </SearchWrapper>
            <ListStyled>
              {this.mapInvitations((searchValue) ? searchResult : pendingInvitations)}
            </ListStyled>
          </Paper>
        </UsersWrapper>
      </PageWrapper>
    )
  }
}

type StateProps = Pick<Props, 'invitations'>
const mapStateToProps = (state: StoreTypes): StateProps => ({
  invitations: selectors.friends.invitations(state),
})

type DispatchProps = Pick<Props, 'onGetInvitations' | 'onAcceptInvitation' | 'onRejectInvitation'>
const mapDispatchToProps = (dispatch): DispatchProps => ({
  onGetInvitations: () => dispatch(Actions.friends.getInvitationsRequest()),
  onAcceptInvitation: (data: ApiTypes.Friends.InvitationAccept) => dispatch(Actions.friends.acceptInvitationRequest(data)),
  onRejectInvitation: (data: ApiTypes.Friends.InvitationReject) => dispatch(Actions.friends.rejectInvitationRequest(data)),
})

export default connect(mapStateToProps, mapDispatchToProps)(Invitations)