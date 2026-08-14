package bot

import (
	"os"
	"path/filepath"

	"github.com/deltachat-bot/deltabot-cli-go/v2/botcli"

	"github.com/omidz4t/portal/internal/config"
	"github.com/omidz4t/portal/internal/dc"
)

// ApplyProfile sets displayname and selfavatar from config on the portal bot account only.
// Persona ghost accounts must keep their own Telegram-derived names.
func ApplyProfile(cli *botcli.BotCli, sess *dc.Session, cfg config.Config) {
	accId, err := sess.FirstConfiguredAccount()
	if err != nil {
		cli.Logger.Error(err)
		return
	}
	if err := applyProfileForAccount(sess, accId, cfg); err != nil {
		cli.GetLogger(accId).Errorf("apply profile: %v", err)
		return
	}
	cli.GetLogger(accId).Infof("profile applied: name=%q image=%q", cfg.Name, cfg.Image)
}

func applyProfileForAccount(sess *dc.Session, accId uint32, cfg config.Config) error {
	if cfg.Name != "" {
		name := cfg.Name
		if err := sess.SetConfig(accId, "displayname", &name); err != nil {
			return err
		}
	}

	if cfg.Image == "" {
		return nil
	}

	path, err := resolveImagePath(cfg.Image)
	if err != nil {
		return err
	}
	return sess.SetConfig(accId, "selfavatar", &path)
}

func resolveImagePath(p string) (string, error) {
	if filepath.IsAbs(p) {
		if _, err := os.Stat(p); err != nil {
			return "", err
		}
		return p, nil
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	return abs, nil
}
