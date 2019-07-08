import * as React from 'react'
import * as Kb from '../../common-adapters'
import * as Styles from '../../styles'
import flags from '../../util/feature-flags'

type Props = {
  airdropIsLive: boolean | null
  bio: string | null
  followThem: boolean | null
  followersCount: number | null
  followingCount: number | null
  followsYou: boolean | null
  fullname: string | null
  inTracker: boolean
  location: string | null
  registeredForAirdrop: boolean | null
  youAreInAirdrop: boolean | null
}

// Here we're using FloatingMenu, but we want to customize the button to match
// Zeplin, so there's a hack -- desktop renders everything as a custom header,
// whereas mobile uses `items` prop as normal.
const _AirdropPopup = p => (
  <Kb.ClickableBox
    ref={p.setAttachmentRef}
    onClick={p.toggleShowingMenu}
    onMouseEnter={p.toggleShowingMenu}
    onMouseLeave={p.toggleShowingMenu}
  >
    <Kb.Icon type="icon-airdrop-star-16" style={styles.star} />
    <Kb.FloatingMenu
      attachTo={p.getAttachmentRef}
      closeOnSelect={false}
      containerStyle={styles.floatingContainer}
      listStyle={styles.floatingContainer}
      backgroundColor={Styles.globalColors.purple}
      textColor={Styles.globalColors.white}
      onHidden={p.toggleShowingMenu}
      visible={true}
      propagateOutsideClicks={true}
      header={{
        title: 'header',
        view: (
          <Kb.Box2
            direction="vertical"
            centerChildren={true}
            fullWidth={true}
            gap="tiny"
            style={{backgroundColor: Styles.globalColors.purple, padding: Styles.globalMargins.small}}
          >
            <Kb.Icon type="icon-airdrop-star-64" style={styles.star} />
            <Kb.Text style={styles.airdropText} type="BodySemibold">
              Join the airdrop
            </Kb.Text>
            <Kb.Text style={styles.airdropText} type="BodySmall">
              Airdropees get free Lumens every month.
            </Kb.Text>
            {!Styles.isMobile && (
              <Kb.Button
                backgroundColor="purple"
                label="Learn more"
                onClick={p.onLearnMore}
                style={styles.learnButton}
              />
            )}
          </Kb.Box2>
        ),
      }}
      items={
        Styles.isMobile
          ? [
              {
                onClick: () => p.onLearnMore,
                title: 'Learn more',
              },
            ]
          : []
      }
    />
  </Kb.ClickableBox>
)
const AirdropPopup = Kb.OverlayParentHOC(_AirdropPopup)

const Bio = (p: Props) => {
  console.warn('bio', p)
  return (
    <Kb.Box2 direction="vertical" fullWidth={true} style={styles.container} centerChildren={true} gap="xtiny">
      <Kb.Box2 direction="horizontal" style={styles.fullNameContainer} gap="tiny">
        <Kb.Text type="BodyBig" lineClamp={p.inTracker ? 1 : undefined} selectable={true}>
          {p.fullname}
        </Kb.Text>
        {flags.airdrop &&
          p.airdropIsLive &&
          p.registeredForAirdrop &&
          (p.youAreInAirdrop ? (
            <Kb.WithTooltip text="Lucky airdropee">
              <Kb.Icon type="icon-airdrop-star-16" style={styles.star} />
            </Kb.WithTooltip>
          ) : (
            <AirdropPopup />
          ))}
      </Kb.Box2>
      {p.followThem && p.followsYou && <Kb.Text type="BodySmall">YOU FOLLOW EACH OTHER</Kb.Text>}
      {p.followersCount !== null && (
        <Kb.Text type="BodySmall">
          <Kb.Text type="BodySmall">
            <Kb.Text type="BodySmall" style={styles.bold}>
              {p.followersCount}
            </Kb.Text>{' '}
            Followers{' '}
          </Kb.Text>
          <Kb.Text type="BodySmall"> · </Kb.Text>
          <Kb.Text type="BodySmall">
            {' '}
            Following{' '}
            <Kb.Text type="BodySmall" style={styles.bold}>
              {p.followingCount}{' '}
            </Kb.Text>
          </Kb.Text>
        </Kb.Text>
      )}
      {!!p.bio && (
        <Kb.Text
          type="Body"
          center={true}
          lineClamp={p.inTracker ? 2 : undefined}
          style={styles.text}
          selectable={true}
        >
          {p.bio}
        </Kb.Text>
      )}
      {!!p.location && (
        <Kb.Text
          type="BodySmall"
          center={true}
          lineClamp={p.inTracker ? 1 : undefined}
          style={styles.text}
          selectable={true}
        >
          {p.location}
        </Kb.Text>
      )}
    </Kb.Box2>
  )
}
const styles = Styles.styleSheetCreate({
  airdropText: Styles.platformStyles({
    common: {color: Styles.globalColors.white},
    isElectron: {textAlign: 'center'},
  }),
  bold: {...Styles.globalStyles.fontBold},
  container: {backgroundColor: Styles.globalColors.white, flexShrink: 0},
  floatingContainer: Styles.platformStyles({
    common: {
      backgroundColor: Styles.globalColors.purple,
    },
    isElectron: {
      maxWidth: 200,
    },
  }),
  fullName: Styles.platformStyles({
    isElectron: {wordBreak: 'break-any'},
  }),
  fullNameContainer: {
    paddingLeft: Styles.globalMargins.mediumLarge,
    paddingRight: Styles.globalMargins.mediumLarge,
  },
  learnButton: {alignSelf: 'center', marginTop: Styles.globalMargins.tiny},
  star: {alignSelf: 'center', marginBottom: Styles.globalMargins.tiny},
  text: Styles.platformStyles({
    common: {
      paddingLeft: Styles.globalMargins.mediumLarge,
      paddingRight: Styles.globalMargins.mediumLarge,
    },
    isElectron: {
      wordBreak: 'break-word',
    },
    isMobile: {
      lineHeight: 21,
    },
  }),
})

export default Bio
