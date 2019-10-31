import * as ProvisionGen from '../../actions/provision-gen'
import * as RouteTreeGen from '../../actions/route-tree-gen'
import CodePage2 from '.'
import * as Container from '../../util/container'
import HiddenString from '../../util/hidden-string'
import * as DevicesConstants from '../../constants/devices'

type OwnProps = Container.RouteProps<{}>

const prov = Container.connect(
  (state: Container.TypedState) => {
    const currentDeviceAlreadyProvisioned = !!state.config.deviceName
    return {
      currentDeviceAlreadyProvisioned,
      // we either have a name for real or we asked on a previous screen
      currentDeviceName:
        (currentDeviceAlreadyProvisioned ? state.config.deviceName : state.provision.deviceName) || '',
      device: DevicesConstants.getDevice(state, state.config.deviceID),
      error: state.provision.error.stringValue(),
      otherDevice: state.provision.codePageOtherDevice,
      textCode: state.provision.codePageIncomingTextCode.stringValue(),
    }
  },
  (dispatch: Container.TypedDispatch) => ({
    onBack: () => dispatch(RouteTreeGen.createNavigateUp()),
    onClose: () => dispatch(ProvisionGen.createCancelProvision()),
    onSubmitTextCode: (code: string) =>
      dispatch(ProvisionGen.createSubmitTextCode({phrase: new HiddenString(code)})),
  }),
  (stateProps, dispatchProps, _: OwnProps) => ({
    currentDevice: stateProps.device,
    currentDeviceAlreadyProvisioned: stateProps.currentDeviceAlreadyProvisioned,
    currentDeviceName: stateProps.currentDeviceName,
    error: stateProps.error,
    onBack: dispatchProps.onBack,
    onClose: dispatchProps.onClose,
    onSubmitTextCode: dispatchProps.onSubmitTextCode,
    otherDevice: stateProps.otherDevice,
    textCode: stateProps.textCode,
  })
)(Container.safeSubmit(['onBack', 'onSubmitTextCode'], ['error'])(CodePage2))
export default prov
