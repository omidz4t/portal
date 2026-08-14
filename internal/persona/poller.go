package persona

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/omidz4t/portal/internal/bridge"
	"github.com/omidz4t/portal/internal/config"
	"github.com/omidz4t/portal/internal/proxy"
	"github.com/omidz4t/portal/internal/safemedia"
	"github.com/omidz4t/portal/internal/store"
	"github.com/omidz4t/portal/internal/usermsg"
)

const personaWorkers = 4

// TGClient is a concrete TelegramAPI for one persona bot.
type TGClient struct {
	api    *tgbotapi.BotAPI
	http   *http.Client
	token  string
	tmpdir string
	log    *zap.SugaredLogger
}

func (c *TGClient) Token() string { return c.token }

func (c *TGClient) GetMeUsername() string {
	return c.api.Self.UserName
}

func (c *TGClient) SendText(chatID int64, text string) error {
	_, err := c.api.Send(tgbotapi.NewMessage(chatID, text))
	return err
}

// DownloadProfilePhoto fetches the user's current TG profile photo if any.
func (c *TGClient) DownloadProfilePhoto(userID int64) (localPath, fileID string, err error) {
	photos, err := c.api.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: userID,
		Limit:  1,
	})
	if err != nil {
		return "", "", err
	}
	if photos.TotalCount == 0 || len(photos.Photos) == 0 || len(photos.Photos[0]) == 0 {
		return "", "", nil
	}
	// Largest size last
	sizes := photos.Photos[0]
	ph := sizes[len(sizes)-1]
	path, err := c.downloadFile(ph.FileID, "avatar.jpg")
	if err != nil {
		return "", "", err
	}
	if err := safemedia.ValidateFile(path, safemedia.RoleAvatar, safemedia.AvatarMaxBytes); err != nil {
		_ = os.Remove(path)
		return "", "", err
	}
	return path, ph.FileID, nil
}

func (c *TGClient) SendMediaFile(chatID int64, path, filename string, asPhoto, asVideo, asAnimation, asDocument bool) error {
	switch {
	case asPhoto:
		p := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(path))
		_, err := c.api.Send(p)
		return err
	case asVideo:
		v := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(path))
		_, err := c.api.Send(v)
		return err
	case asAnimation:
		a := tgbotapi.NewAnimation(chatID, tgbotapi.FilePath(path))
		_, err := c.api.Send(a)
		return err
	default:
		d := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(path))
		if filename != "" {
			// library uses path base; caption unused
			_ = filename
		}
		_, err := c.api.Send(d)
		return err
	}
}

// NewTGClient authenticates a BotFather token.
func NewTGClient(token string, cfg config.Config, log *zap.SugaredLogger) (*TGClient, error) {
	pc := cfg.TelegramProxy()
	httpClient, err := proxy.HTTPClient(pc, 120*time.Second)
	if err != nil {
		return nil, err
	}
	api, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		return nil, fmt.Errorf("telegram auth: %w", err)
	}
	tmpdir := filepath.Join(cfg.Folder, "tg-cache", "persona")
	if err := os.MkdirAll(tmpdir, 0o755); err != nil {
		return nil, err
	}
	return &TGClient{
		api:    api,
		http:   httpClient,
		token:  token,
		tmpdir: tmpdir,
		log:    log,
	}, nil
}

// StartPoller builds a BotRuntime that long-polls and bridges into Manager.
func (m *Manager) StartPoller(bot store.PersonaBot) (*BotRuntime, error) {
	client, err := NewTGClient(bot.BotToken, m.cfg, m.log.With("persona_bot", bot.ID, "username", bot.BotUsername))
	if err != nil {
		return nil, err
	}
	// Refresh username if empty
	if bot.BotUsername == "" {
		bot.BotUsername = client.api.Self.UserName
	}
	rt := &BotRuntime{
		Bot:     bot,
		API:     client,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	go m.pollLoop(rt, client)
	return rt, nil
}

func (m *Manager) pollLoop(rt *BotRuntime, c *TGClient) {
	defer close(rt.stopped)
	sem := make(chan struct{}, personaWorkers)
	offset := 0
	c.log.Infof("persona poll started @%s", c.api.Self.UserName)

	for {
		select {
		case <-rt.stop:
			c.log.Info("persona poll stopped")
			return
		default:
		}

		params := tgbotapi.Params{}
		params.AddNonZero("timeout", 50)
		params.AddNonZero("offset", offset)
		// message = DMs + groups (if BotFather privacy mode is disabled for the bot)
		// my_chat_member = detect being added/removed from groups
		params["allowed_updates"] = `["message","edited_message","my_chat_member"]`

		resp, err := c.api.MakeRequest("getUpdates", params)
		if err != nil {
			// stop requested during long poll
			select {
			case <-rt.stop:
				return
			default:
			}
			c.log.Errorf("getUpdates: %v", err)
			// Unauthorized → mark error
			if strings.Contains(err.Error(), "Unauthorized") {
				_ = m.store.SetPersonaBotStatus(rt.Bot.ID, store.PersonaBotError)
				m.dropPoller(rt.Bot.ID)
				return
			}
			time.Sleep(2 * time.Second)
			continue
		}

		var rawUpdates []json.RawMessage
		if err := json.Unmarshal(resp.Result, &rawUpdates); err != nil {
			time.Sleep(time.Second)
			continue
		}
		for _, raw := range rawUpdates {
			var meta struct {
				UpdateID      int             `json:"update_id"`
				Message       json.RawMessage `json:"message"`
				EditedMessage json.RawMessage `json:"edited_message"`
				MyChatMember  json.RawMessage `json:"my_chat_member"`
			}
			if err := json.Unmarshal(raw, &meta); err != nil {
				continue
			}
			if meta.UpdateID >= offset {
				offset = meta.UpdateID + 1
			}
			// Log when bot is added to a group (helps debug privacy / membership).
			if len(meta.MyChatMember) > 0 && string(meta.MyChatMember) != "null" {
				m.logMyChatMember(c, meta.MyChatMember)
			}
			msgRaw := meta.Message
			if len(msgRaw) == 0 || string(msgRaw) == "null" {
				msgRaw = meta.EditedMessage
			}
			if len(msgRaw) == 0 || string(msgRaw) == "null" {
				continue
			}
			var msg tgbotapi.Message
			if err := json.Unmarshal(msgRaw, &msg); err != nil {
				continue
			}
			select {
			case <-rt.stop:
				return
			default:
			}
			sem <- struct{}{}
			go func(msg tgbotapi.Message) {
				defer func() { <-sem }()
				if err := m.handlePersonaMessage(rt, c, &msg); err != nil {
					c.log.Errorf("handle: %v", err)
					// In groups, only reply when we can (avoid spam); always try once.
					_, _ = c.api.Send(tgbotapi.NewMessage(msg.Chat.ID, usermsg.BridgeFailed))
				}
			}(msg)
		}
	}
}

func (m *Manager) logMyChatMember(c *TGClient, raw json.RawMessage) {
	var u struct {
		Chat struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"chat"`
		NewChatMember struct {
			Status string `json:"status"`
			User   struct {
				ID       int64  `json:"id"`
				IsBot    bool   `json:"is_bot"`
				Username string `json:"username"`
			} `json:"user"`
		} `json:"new_chat_member"`
	}
	if err := json.Unmarshal(raw, &u); err != nil {
		return
	}
	c.log.Infof("chat_member chat=%d type=%s title=%q status=%s bot=@%s",
		u.Chat.ID, u.Chat.Type, u.Chat.Title, u.NewChatMember.Status, u.NewChatMember.User.Username)
	if IsTelegramGroup(u.Chat.ID, u.Chat.Type) &&
		(u.NewChatMember.Status == "member" || u.NewChatMember.Status == "administrator") {
		c.log.Infof("bot joined group %q — for full message bridging, disable privacy mode in BotFather: /setprivacy → Disable",
			u.Chat.Title)
	}
}

func (m *Manager) handlePersonaMessage(rt *BotRuntime, c *TGClient, msg *tgbotapi.Message) error {
	if msg.From == nil || msg.From.IsBot {
		return nil
	}
	bot := rt.Bot
	// Commands on persona bot
	if msg.Text != "" && strings.HasPrefix(msg.Text, "/") {
		cmd := strings.Fields(msg.Text)[0]
		if i := strings.Index(cmd, "@"); i > 0 {
			cmd = cmd[:i]
		}
		switch strings.ToLower(cmd) {
		case "/start", "/help":
			help := "This bot bridges to the owner's Delta Chat.\n" +
				"DMs and group messages appear as your contact there.\n\n" +
				"Groups: add this bot to a group. The bot owner must disable " +
				"privacy mode in @BotFather (/setprivacy → Disable) so the bot " +
				"receives all group messages (not only @mentions)."
			_, err := c.api.Send(tgbotapi.NewMessage(msg.Chat.ID, help))
			return err
		default:
			// ignore other commands
		}
	}

	from := TGUser{
		ID:          msg.From.ID,
		Username:    msg.From.UserName,
		DisplayName: strings.TrimSpace(msg.From.FirstName + " " + msg.From.LastName),
	}
	// Prefer chat id sign: Telegram group/supergroup ids are always negative.
	isGroup := IsTelegramGroup(msg.Chat.ID, msg.Chat.Type) ||
		msg.Chat.IsGroup() || msg.Chat.IsSuperGroup()
	title := msg.Chat.Title
	if isGroup {
		if !m.cfg.Persona.AllowGroups {
			return nil
		}
		c.log.Infof("group message chat=%d type=%s title=%q from=%d",
			msg.Chat.ID, msg.Chat.Type, title, msg.From.ID)
	}

	in := Incoming{
		Bot:        &bot,
		From:       from,
		ChatID:     msg.Chat.ID,
		IsGroup:    isGroup,
		GroupTitle: title,
		Avatar:     c,
	}

	// Text
	if msg.Text != "" && !strings.HasPrefix(msg.Text, "/") {
		in.Text = msg.Text
		return m.BridgeToDelta(in)
	}

	// Media downloads
	var fileID, name string
	var kind bridge.Kind
	switch {
	case msg.Photo != nil && len(msg.Photo) > 0:
		ph := msg.Photo[len(msg.Photo)-1]
		fileID, name, kind = ph.FileID, "image.jpg", bridge.KindImage
		in.Viewtype = bridge.ViewImage()
		in.Text = strings.TrimSpace(msg.Caption)
	case msg.Video != nil:
		fileID, name, kind = msg.Video.FileID, "video.mp4", bridge.KindVideo
		in.Viewtype = bridge.ViewVideo()
		in.Text = strings.TrimSpace(msg.Caption)
	case msg.Animation != nil:
		fileID, name, kind = msg.Animation.FileID, "video.mp4", bridge.KindGif
		in.Viewtype = bridge.ViewVideo()
		in.Text = strings.TrimSpace(msg.Caption)
	case msg.Sticker != nil:
		fileID = msg.Sticker.FileID
		name = "sticker.webp"
		kind = bridge.KindSticker
		in.Viewtype = bridge.ViewSticker()
	case msg.Document != nil:
		fileID = msg.Document.FileID
		name = "file.bin"
		if msg.Document.FileName != "" {
			ext := filepath.Ext(msg.Document.FileName)
			if ext != "" {
				name = "file" + strings.ToLower(ext)
			}
		}
		kind = bridge.KindFromFilename(name)
		if kind == "" {
			kind = bridge.KindImage
		}
		in.Viewtype = bridge.ViewFile()
		in.Text = strings.TrimSpace(msg.Caption)
	default:
		return nil
	}
	_ = kind

	path, err := c.downloadFile(fileID, name)
	if err != nil {
		return err
	}
	defer os.Remove(path)
	role := safemedia.RoleFromKind(string(kind))
	if msg.Document != nil {
		role = safemedia.RoleFile
	}
	if err := safemedia.ValidateFile(path, role, safemedia.DefaultMaxBytes); err != nil {
		return err
	}
	in.FilePath = path
	in.FileName = name
	return m.BridgeToDelta(in)
}

func (c *TGClient) downloadFile(fileID, preferredName string) (string, error) {
	f, err := c.api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", err
	}
	url := f.Link(c.token)
	ext := filepath.Ext(preferredName)
	if ext == "" {
		ext = ".bin"
	}
	out := filepath.Join(c.tmpdir, fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))
	resp, err := c.http.Get(url) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download: HTTP %s", resp.Status)
	}
	w, err := os.Create(out)
	if err != nil {
		return "", err
	}
	_, copyErr := safemedia.CopyLimited(w, resp.Body, safemedia.DefaultMaxBytes)
	_ = w.Close()
	if copyErr != nil {
		_ = os.Remove(out)
		return "", copyErr
	}
	return out, nil
}

func sendVia(api TelegramAPI, tgChatID int64, text, path, name string) error {
	if path != "" {
		ext := strings.ToLower(filepath.Ext(name))
		asPhoto := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
		asVideo := ext == ".mp4" || ext == ".webm" || ext == ".mov"
		asAnim := ext == ".gif"
		if err := api.SendMediaFile(tgChatID, path, name, asPhoto, asVideo, asAnim, !asPhoto && !asVideo && !asAnim); err != nil {
			return err
		}
		if text != "" {
			return api.SendText(tgChatID, text)
		}
		return nil
	}
	if text == "" {
		return nil
	}
	return api.SendText(tgChatID, text)
}
