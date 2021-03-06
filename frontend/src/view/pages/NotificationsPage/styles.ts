import styled from 'styled-components'
import Container from '@material-ui/core/Container'
import Paper from '@material-ui/core/Paper'
import { Link } from 'react-router-dom'

export const ContainerStyled = styled(Container)`
  && {
    margin-top: 90px;
  }
`

export const NotificationsWrapper = styled(Paper)`
  padding: 15px;
  min-height: 50vh;
  position: relative;
  padding-bottom: 80px;
`

export const Header = styled.header`
  display: flex;
  justify-content: space-between;
  align-items: center;

  @media (max-width: 600px) {
    margin-bottom: 20px;
  }
` 

export const Footer = styled.footer`
  display: flex;
  justify-content: flex-end;
  align-items: center;
  position: absolute;
  width: 100%;
  bottom: 15px;
  left: 0px;
  padding: 0 15px;
` 

export const Title = styled.h3`
  font-size: 18px;
  margin: 0;
`

export const ListWrapper = styled.div`
  margin-top: 10px;
`

export const ListIten = styled.div`
  margin-bottom: 10px;
  display: flex;
  flex-wrap: wrap;

  @media (max-width: 600px) {
    padding-bottom: 10px;
    border-bottom: 1px solid #ccc;
  }
`

export const ListDate = styled.span`
  width: 180px;

  @media (max-width: 600px) {
    width: 100%;
  }
`

export const ListText = styled.p`
  margin: 0;
  display: flex;
  align-items: center;

  @media (max-width: 600px) {
    width: 100%;
  }
`

export const ListLink = styled(Link)`
  margin-left: 7px;
  text-decoration: underline;
  cursor: pointer;
  color: #3f51b5;

  &:hover {
    text-decoration: none;
  }
`

export const ButtonsWrapper = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 20px 0;
`
