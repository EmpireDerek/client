/* @flow */
/*eslint-disable react/prop-types */ // Since we're using flow types for props

import React, {Component} from '../base-react'
import Render from './index.render'
import {connect} from '../base-redux'
import {bindActionCreators} from 'redux'

import * as favoriteAction from '../actions/favorite'

import {remote, shell} from 'electron'
import {ipcRenderer} from 'electron'

import {kbfsPath} from '../constants/platform'

export type MenubarProps = {
  username: ?string,
  folders: ?string,
  favoriteList: () => void,
  debug: ?boolean
}

class Menubar extends Component {
  props: MenubarProps;

  checkForFolders () {
    if (this.props.username && !this.props.folders) {
      this.props.favoriteList()
    }
  }

  componentDidMount () {
    this.checkForFolders()
  }

  componentDidUpdate () {
    this.checkForFolders()
  }

  render () {
    const openingMessage = this.props.username ? 'Be sure to drink your Ovaltine.' : "Looks like you aren't logged in. Try running `keybase login`"

    const closeMenubar = () => ipcRenderer.send('closeMenubar')

    // TODO change this to /keybase when all our mount points are `/keybase/...`
    // TODO Support linux
    const openKBFS = () => { shell.openItem(kbfsPath); closeMenubar() }
    const openKBFSPublic = () => { shell.openItem(`${kbfsPath}/public/${this.props.username}`); closeMenubar() }
    const openKBFSPrivate = () => { shell.openItem(`${kbfsPath}/private/${this.props.username}`); closeMenubar() }

    const showMain = () => { ipcRenderer.send('showMain'); closeMenubar() }
    const showHelp = () => { ipcRenderer.send('showHelp'); closeMenubar() }
    const quit = () => remote.app.emit('destroy')

    const openingButtonInfo = this.props.username && {text: 'Get Started', onClick: showHelp}

    // TODO (pull this debug from somewhere
    return <Render {...{username: this.props.username, openingMessage: openingMessage, debug: !!this.props.debug, openingButtonInfo, openKBFS, openKBFSPublic, openKBFSPrivate, showMain, showHelp, quit}}/>
  }
}

export default connect(
  state => ({
    username: state.config && state.config.status && state.config.status.user && state.config.status.user.username,
    folders: state.favorite.folders
  }),
  dispatch => bindActionCreators(favoriteAction, dispatch)
)(Menubar)
