package bot

import (
	"github.com/deltachat-bot/deltabot-cli-go/v2/botcli"

	"github.com/omidz4t/portal/internal/dc"
)

// NotifyInviteContact joins/opens a chat from INVITE_URL and sends bootMessage
// on the portal account only (not persona ghost accounts).
func NotifyInviteContact(cli *botcli.BotCli, sess *dc.Session, inviteURL, bootMessage string) {
	if inviteURL == "" {
		cli.Logger.Info("INVITE_URL not set; skipping boot notification")
		return
	}
	if bootMessage == "" {
		return
	}

	accId, err := sess.FirstConfiguredAccount()
	if err != nil {
		cli.Logger.Error(err)
		return
	}

	go func(accId uint32) {
		chatID, err := sess.OpenChatFromInvite(accId, inviteURL)
		if err != nil {
			cli.GetLogger(accId).Errorf("boot notify failed: %v", err)
			return
		}
		if err := sess.SendTextWithRetry(accId, chatID, bootMessage, 40); err != nil {
			cli.GetLogger(accId).Errorf("boot notify failed: %v", err)
			return
		}
		cli.GetLogger(accId).Infof("sent boot message to invite contact")
	}(accId)
}
