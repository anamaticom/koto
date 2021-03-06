import styled from 'styled-components'
import FormControl from '@material-ui/core/FormControl'
import Button from '@material-ui/core/Button'
import Container from '@material-ui/core/Container'
import { Link } from 'react-router-dom'

export const FormWrapper = styled.form`
  width: 280px;
  margin: 0 auto;
`

export const FormControlStyled = styled(FormControl)`
  && {
    margin: 0 0 15px; 
    width: 100%;
  }
`

export const ButtonStyled = styled(Button)`
  && {
    width: 100%;
    height: 42px;
  }
`

export const ContainerStyled = styled(Container)`
  && {
    display: flex;
    flex-wrap: wrap;
    align-content: center;
    justify-content: center;
    min-height: 100vh;
  }
`

export const Header = styled.div`
  padding-bottom: 30px;
  text-align: center;
`

export const Logo = styled.img`
  max-height: 55px;
  position: relative;
  margin-bottom: 20px;
`

export const CheckBoxWrapper = styled.div`
  margin: -5px 0 10px -10px;
  display: flex;
  align-items: center;
`

export const CheckBoxLabel = styled(Link)`
  font-size: 15px;
  line-height: 1;
  color: #1976d2;
  text-decoration: none;
  display: block;

  &:hover {
    text-decoration: underline;
  }

`
