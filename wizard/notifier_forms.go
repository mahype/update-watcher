package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mahype/update-watcher/config"
)

// addFuncs maps notifier types to their add-configuration functions.
var addFuncs = map[string]func(cfg *config.Config) error{
	"slack":      addSlack,
	"ntfy":       addNtfy,
	"webhook":    addWebhook,
	"discord":    addDiscord,
	"telegram":   addTelegram,
	"teams":      addTeams,
	"email":      addEmail,
	"pushover":   addPushover,
	"gotify":     addGotify,
	"googlechat":  addGoogleChat,
	"matrix":      addMatrix,
	"mattermost":  addMattermost,
	"rocketchat":  addRocketChat,
	"pagerduty":   addPagerDuty,
	"pushbullet":  addPushbullet,
}

// editFuncs maps notifier types to their edit-configuration functions.
var editFuncs = map[string]func(cfg *config.Config, existing *config.NotifierConfig) error{
	"slack":      editSlack,
	"ntfy":       editNtfy,
	"webhook":    editWebhook,
	"discord":    editDiscord,
	"telegram":   editTelegram,
	"teams":      editTeams,
	"email":      editEmail,
	"pushover":   editPushover,
	"gotify":     editGotify,
	"googlechat":  editGoogleChat,
	"matrix":      editMatrix,
	"mattermost":  editMattermost,
	"rocketchat":  editRocketChat,
	"pagerduty":   editPagerDuty,
	"pushbullet":  editPushbullet,
}

// --- Slack ---

func addSlack(cfg *config.Config) error {
	var webhookURL string
	var mentionOnSecurity bool

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Slack Webhook URL").
				Description("Create one at https://api.slack.com/messaging/webhooks").
				Value(&webhookURL),
			huh.NewConfirm().
				Title("Mention @channel for security updates?").
				Value(&mentionOnSecurity),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"webhook_url": webhookURL,
		"use_emoji":   true,
	}
	if mentionOnSecurity {
		options["mention_on_security"] = "@channel"
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "slack",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Slack notifier added.")
	return nil
}

func editSlack(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	webhookURL := opts.GetString("webhook_url", "")
	mentionOnSecurity := opts.GetString("mention_on_security", "") != ""

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Slack Webhook URL").
				Description("Leave unchanged or enter a new URL").
				Value(&webhookURL),
			huh.NewConfirm().
				Title("Mention @channel for security updates?").
				Value(&mentionOnSecurity),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["webhook_url"] = webhookURL
	if mentionOnSecurity {
		existing.Options["mention_on_security"] = "@channel"
	} else {
		delete(existing.Options, "mention_on_security")
	}

	fmt.Println("  Slack settings updated.")
	return nil
}

// --- ntfy ---

func addNtfy(cfg *config.Config) error {
	serverURL := "https://ntfy.sh"
	var topic, token string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Server URL").
				Description("Default: https://ntfy.sh (or your self-hosted instance)").
				Value(&serverURL),
			huh.NewInput().
				Title("Topic").
				Description("The ntfy topic to publish to (required)").
				Value(&topic),
			huh.NewInput().
				Title("Auth Token").
				Description("Optional access token (leave empty for public topics)").
				Value(&token),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"server_url": serverURL,
		"topic":      topic,
		"priority":   "default",
	}
	if token != "" {
		options["token"] = token
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "ntfy",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  ntfy notifier added.")
	return nil
}

func editNtfy(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	serverURL := opts.GetString("server_url", "https://ntfy.sh")
	topic := opts.GetString("topic", "")
	token := opts.GetString("token", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Server URL").
				Value(&serverURL),
			huh.NewInput().
				Title("Topic").
				Value(&topic),
			huh.NewInput().
				Title("Auth Token").
				Description("Leave empty for public topics").
				Value(&token),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["server_url"] = serverURL
	existing.Options["topic"] = topic
	if token != "" {
		existing.Options["token"] = token
	} else {
		delete(existing.Options, "token")
	}

	fmt.Println("  ntfy settings updated.")
	return nil
}

// --- Webhook ---

func addWebhook(cfg *config.Config) error {
	var url, authHeader string
	method := "POST"

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Webhook URL").
				Description("The HTTP endpoint to send JSON payloads to (required)").
				Value(&url),
			huh.NewInput().
				Title("HTTP Method").
				Description("Default: POST").
				Value(&method),
			huh.NewInput().
				Title("Authorization Header").
				Description("e.g. 'Bearer my-secret-token' (leave empty for none)").
				Value(&authHeader),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"url":          url,
		"method":       method,
		"content_type": "application/json",
	}
	if authHeader != "" {
		options["auth_header"] = authHeader
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "webhook",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Webhook notifier added.")
	return nil
}

func editWebhook(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	url := opts.GetString("url", "")
	method := opts.GetString("method", "POST")
	authHeader := opts.GetString("auth_header", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Webhook URL").
				Value(&url),
			huh.NewInput().
				Title("HTTP Method").
				Value(&method),
			huh.NewInput().
				Title("Authorization Header").
				Description("Leave empty for none").
				Value(&authHeader),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["url"] = url
	existing.Options["method"] = method
	if authHeader != "" {
		existing.Options["auth_header"] = authHeader
	} else {
		delete(existing.Options, "auth_header")
	}

	fmt.Println("  Webhook settings updated.")
	return nil
}

// --- Discord ---

func addDiscord(cfg *config.Config) error {
	var webhookURL, username, mentionRole string
	username = "Update Watcher"

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Discord Webhook URL").
				Description("Server Settings > Integrations > Webhooks (required)").
				Value(&webhookURL),
			huh.NewInput().
				Title("Bot Username").
				Description("Display name for the webhook bot").
				Value(&username),
			huh.NewInput().
				Title("Mention Role ID").
				Description("Discord role ID to mention on security updates (leave empty for none)").
				Value(&mentionRole),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"webhook_url": webhookURL,
		"username":    username,
	}
	if mentionRole != "" {
		options["mention_role"] = mentionRole
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "discord",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Discord notifier added.")
	return nil
}

func editDiscord(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	webhookURL := opts.GetString("webhook_url", "")
	username := opts.GetString("username", "Update Watcher")
	mentionRole := opts.GetString("mention_role", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Discord Webhook URL").
				Value(&webhookURL),
			huh.NewInput().
				Title("Bot Username").
				Value(&username),
			huh.NewInput().
				Title("Mention Role ID").
				Description("Leave empty for none").
				Value(&mentionRole),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["webhook_url"] = webhookURL
	existing.Options["username"] = username
	if mentionRole != "" {
		existing.Options["mention_role"] = mentionRole
	} else {
		delete(existing.Options, "mention_role")
	}

	fmt.Println("  Discord settings updated.")
	return nil
}

// --- Telegram ---

func addTelegram(cfg *config.Config) error {
	var botToken, chatID string
	var disableNotification bool

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Bot Token").
				Description("From @BotFather, e.g. 123456:ABC-DEF... (required)").
				Value(&botToken),
			huh.NewInput().
				Title("Chat ID").
				Description("User/group/channel ID to send messages to (required)").
				Value(&chatID),
			huh.NewConfirm().
				Title("Send silent notifications?").
				Description("Notifications without sound").
				Value(&disableNotification),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"bot_token":  botToken,
		"chat_id":    chatID,
		"parse_mode": "HTML",
	}
	if disableNotification {
		options["disable_notification"] = true
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "telegram",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Telegram notifier added.")
	return nil
}

func editTelegram(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	botToken := opts.GetString("bot_token", "")
	chatID := opts.GetString("chat_id", "")
	disableNotification := opts.GetBool("disable_notification", false)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Bot Token").
				Value(&botToken),
			huh.NewInput().
				Title("Chat ID").
				Value(&chatID),
			huh.NewConfirm().
				Title("Send silent notifications?").
				Value(&disableNotification),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["bot_token"] = botToken
	existing.Options["chat_id"] = chatID
	if disableNotification {
		existing.Options["disable_notification"] = true
	} else {
		delete(existing.Options, "disable_notification")
	}

	fmt.Println("  Telegram settings updated.")
	return nil
}

// --- Teams ---

func addTeams(cfg *config.Config) error {
	var webhookURL string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Teams Webhook URL").
				Description("Power Automate Workflow webhook URL (required)").
				Value(&webhookURL),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"webhook_url": webhookURL,
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "teams",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Microsoft Teams notifier added.")
	return nil
}

func editTeams(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	webhookURL := opts.GetString("webhook_url", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Teams Webhook URL").
				Value(&webhookURL),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["webhook_url"] = webhookURL

	fmt.Println("  Microsoft Teams settings updated.")
	return nil
}

// --- Email ---

func addEmail(cfg *config.Config) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	var username, password, from, toStr string
	useTLS := true

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("SMTP Host").
				Description("e.g. smtp.gmail.com").
				Value(&smtpHost),
			huh.NewInput().
				Title("SMTP Port").
				Description("587 for STARTTLS, 465 for SSL").
				Value(&smtpPort),
			huh.NewInput().
				Title("Username").
				Value(&username),
			huh.NewInput().
				Title("Password").
				Description("App password or SMTP password").
				EchoMode(huh.EchoModePassword).
				Value(&password),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("From Address").
				Description("Sender email address").
				Value(&from),
			huh.NewInput().
				Title("To Addresses").
				Description("Comma-separated recipient addresses").
				Value(&toStr),
			huh.NewConfirm().
				Title("Use TLS (STARTTLS)?").
				Value(&useTLS),
		),
	).Run()
	if err != nil {
		return nil
	}

	// Parse port
	var port int
	fmt.Sscanf(smtpPort, "%d", &port)
	if port == 0 {
		port = 587
	}

	// Parse recipients
	var to []interface{}
	for _, addr := range strings.Split(toStr, ",") {
		if t := strings.TrimSpace(addr); t != "" {
			to = append(to, t)
		}
	}

	options := map[string]interface{}{
		"smtp_host": smtpHost,
		"smtp_port": port,
		"username":  username,
		"password":  password,
		"from":      from,
		"to":        to,
		"tls":       useTLS,
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "email",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  E-Mail notifier added.")
	return nil
}

func editEmail(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	smtpHost := opts.GetString("smtp_host", "")
	smtpPort := fmt.Sprintf("%d", getIntOption(existing.Options, "smtp_port", 587))
	username := opts.GetString("username", "")
	password := opts.GetString("password", "")
	from := opts.GetString("from", "")
	toSlice := opts.GetStringSlice("to", nil)
	toStr := strings.Join(toSlice, ", ")
	useTLS := opts.GetBool("tls", true)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("SMTP Host").
				Value(&smtpHost),
			huh.NewInput().
				Title("SMTP Port").
				Value(&smtpPort),
			huh.NewInput().
				Title("Username").
				Value(&username),
			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&password),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("From Address").
				Value(&from),
			huh.NewInput().
				Title("To Addresses").
				Description("Comma-separated").
				Value(&toStr),
			huh.NewConfirm().
				Title("Use TLS (STARTTLS)?").
				Value(&useTLS),
		),
	).Run()
	if err != nil {
		return nil
	}

	var port int
	fmt.Sscanf(smtpPort, "%d", &port)
	if port == 0 {
		port = 587
	}

	var to []interface{}
	for _, addr := range strings.Split(toStr, ",") {
		if t := strings.TrimSpace(addr); t != "" {
			to = append(to, t)
		}
	}

	existing.Options["smtp_host"] = smtpHost
	existing.Options["smtp_port"] = port
	existing.Options["username"] = username
	existing.Options["password"] = password
	existing.Options["from"] = from
	existing.Options["to"] = to
	existing.Options["tls"] = useTLS

	fmt.Println("  E-Mail settings updated.")
	return nil
}

func getIntOption(options map[string]interface{}, key string, defaultVal int) int {
	if v, ok := options[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return defaultVal
}

// --- Pushover ---

func addPushover(cfg *config.Config) error {
	var appToken, userKey, device, sound string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Application Token").
				Description("From https://pushover.net/apps (required)").
				Value(&appToken),
			huh.NewInput().
				Title("User/Group Key").
				Description("Your Pushover user or group key (required)").
				Value(&userKey),
			huh.NewInput().
				Title("Device").
				Description("Send to specific device only (leave empty for all)").
				Value(&device),
			huh.NewInput().
				Title("Sound").
				Description("Notification sound (leave empty for default)").
				Value(&sound),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"app_token": appToken,
		"user_key":  userKey,
	}
	if device != "" {
		options["device"] = device
	}
	if sound != "" {
		options["sound"] = sound
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "pushover",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Pushover notifier added.")
	return nil
}

func editPushover(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	appToken := opts.GetString("app_token", "")
	userKey := opts.GetString("user_key", "")
	device := opts.GetString("device", "")
	sound := opts.GetString("sound", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Application Token").
				Value(&appToken),
			huh.NewInput().
				Title("User/Group Key").
				Value(&userKey),
			huh.NewInput().
				Title("Device").
				Description("Leave empty for all devices").
				Value(&device),
			huh.NewInput().
				Title("Sound").
				Description("Leave empty for default").
				Value(&sound),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["app_token"] = appToken
	existing.Options["user_key"] = userKey
	if device != "" {
		existing.Options["device"] = device
	} else {
		delete(existing.Options, "device")
	}
	if sound != "" {
		existing.Options["sound"] = sound
	} else {
		delete(existing.Options, "sound")
	}

	fmt.Println("  Pushover settings updated.")
	return nil
}

// --- Gotify ---

func addGotify(cfg *config.Config) error {
	var serverURL, token string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Gotify Server URL").
				Description("e.g. https://gotify.example.com (required)").
				Value(&serverURL),
			huh.NewInput().
				Title("Application Token").
				Description("From Gotify web UI > Apps (required)").
				Value(&token),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"server_url": serverURL,
		"token":      token,
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "gotify",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Gotify notifier added.")
	return nil
}

func editGotify(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	serverURL := opts.GetString("server_url", "")
	token := opts.GetString("token", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Gotify Server URL").
				Value(&serverURL),
			huh.NewInput().
				Title("Application Token").
				Value(&token),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["server_url"] = serverURL
	existing.Options["token"] = token

	fmt.Println("  Gotify settings updated.")
	return nil
}

// --- Google Chat ---

func addGoogleChat(cfg *config.Config) error {
	var webhookURL, threadKey string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Google Chat Webhook URL").
				Description("Space settings > Apps & integrations > Webhooks (required)").
				Value(&webhookURL),
			huh.NewInput().
				Title("Thread Key").
				Description("Group messages in a thread (leave empty for new messages)").
				Value(&threadKey),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"webhook_url": webhookURL,
	}
	if threadKey != "" {
		options["thread_key"] = threadKey
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "googlechat",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Google Chat notifier added.")
	return nil
}

func editGoogleChat(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	webhookURL := opts.GetString("webhook_url", "")
	threadKey := opts.GetString("thread_key", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Google Chat Webhook URL").
				Value(&webhookURL),
			huh.NewInput().
				Title("Thread Key").
				Description("Leave empty for new messages").
				Value(&threadKey),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["webhook_url"] = webhookURL
	if threadKey != "" {
		existing.Options["thread_key"] = threadKey
	} else {
		delete(existing.Options, "thread_key")
	}

	fmt.Println("  Google Chat settings updated.")
	return nil
}

// --- Matrix ---

func addMatrix(cfg *config.Config) error {
	var homeserver, accessToken, roomID string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Homeserver URL").
				Description("e.g. https://matrix.org (required)").
				Value(&homeserver),
			huh.NewInput().
				Title("Access Token").
				Description("Bot access token (required)").
				Value(&accessToken),
			huh.NewInput().
				Title("Room ID").
				Description("e.g. !abc123:matrix.org (required)").
				Value(&roomID),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"homeserver":   homeserver,
		"access_token": accessToken,
		"room_id":      roomID,
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "matrix",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Matrix notifier added.")
	return nil
}

func editMatrix(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	homeserver := opts.GetString("homeserver", "")
	accessToken := opts.GetString("access_token", "")
	roomID := opts.GetString("room_id", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Homeserver URL").
				Value(&homeserver),
			huh.NewInput().
				Title("Access Token").
				Value(&accessToken),
			huh.NewInput().
				Title("Room ID").
				Value(&roomID),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["homeserver"] = homeserver
	existing.Options["access_token"] = accessToken
	existing.Options["room_id"] = roomID

	fmt.Println("  Matrix settings updated.")
	return nil
}

// --- Mattermost ---

func addMattermost(cfg *config.Config) error {
	var webhookURL, channel, iconURL string
	username := "Update Watcher"

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Mattermost Webhook URL").
				Description("Main Menu > Integrations > Incoming Webhooks (required)").
				Value(&webhookURL),
			huh.NewInput().
				Title("Channel").
				Description("Override channel (leave empty for webhook default)").
				Value(&channel),
			huh.NewInput().
				Title("Username").
				Description("Bot display name").
				Value(&username),
			huh.NewInput().
				Title("Icon URL").
				Description("Bot avatar URL (leave empty for default)").
				Value(&iconURL),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"webhook_url": webhookURL,
		"username":    username,
	}
	if channel != "" {
		options["channel"] = channel
	}
	if iconURL != "" {
		options["icon_url"] = iconURL
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "mattermost",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Mattermost notifier added.")
	return nil
}

func editMattermost(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	webhookURL := opts.GetString("webhook_url", "")
	channel := opts.GetString("channel", "")
	username := opts.GetString("username", "Update Watcher")
	iconURL := opts.GetString("icon_url", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Mattermost Webhook URL").
				Value(&webhookURL),
			huh.NewInput().
				Title("Channel").
				Description("Leave empty for webhook default").
				Value(&channel),
			huh.NewInput().
				Title("Username").
				Value(&username),
			huh.NewInput().
				Title("Icon URL").
				Description("Leave empty for default").
				Value(&iconURL),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["webhook_url"] = webhookURL
	existing.Options["username"] = username
	if channel != "" {
		existing.Options["channel"] = channel
	} else {
		delete(existing.Options, "channel")
	}
	if iconURL != "" {
		existing.Options["icon_url"] = iconURL
	} else {
		delete(existing.Options, "icon_url")
	}

	fmt.Println("  Mattermost settings updated.")
	return nil
}

// --- Rocket.Chat ---

func addRocketChat(cfg *config.Config) error {
	var webhookURL, channel string
	username := "Update Watcher"

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Rocket.Chat Webhook URL").
				Description("Administration > Integrations > Incoming (required)").
				Value(&webhookURL),
			huh.NewInput().
				Title("Channel").
				Description("e.g. #general (leave empty for webhook default)").
				Value(&channel),
			huh.NewInput().
				Title("Username").
				Description("Bot display name").
				Value(&username),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"webhook_url": webhookURL,
		"username":    username,
	}
	if channel != "" {
		options["channel"] = channel
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "rocketchat",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Rocket.Chat notifier added.")
	return nil
}

func editRocketChat(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	webhookURL := opts.GetString("webhook_url", "")
	channel := opts.GetString("channel", "")
	username := opts.GetString("username", "Update Watcher")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Rocket.Chat Webhook URL").
				Value(&webhookURL),
			huh.NewInput().
				Title("Channel").
				Description("Leave empty for webhook default").
				Value(&channel),
			huh.NewInput().
				Title("Username").
				Value(&username),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["webhook_url"] = webhookURL
	existing.Options["username"] = username
	if channel != "" {
		existing.Options["channel"] = channel
	} else {
		delete(existing.Options, "channel")
	}

	fmt.Println("  Rocket.Chat settings updated.")
	return nil
}

// --- PagerDuty ---

func addPagerDuty(cfg *config.Config) error {
	var routingKey string
	severity := "warning"

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Routing Key").
				Description("Events API v2 integration key (required)").
				Value(&routingKey),
			huh.NewSelect[string]().
				Title("Default Severity").
				Options(
					huh.NewOption("Info", "info"),
					huh.NewOption("Warning", "warning"),
					huh.NewOption("Error", "error"),
					huh.NewOption("Critical", "critical"),
				).
				Value(&severity),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"routing_key": routingKey,
		"severity":    severity,
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "pagerduty",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  PagerDuty notifier added.")
	return nil
}

func editPagerDuty(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	routingKey := opts.GetString("routing_key", "")
	severity := opts.GetString("severity", "warning")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Routing Key").
				Value(&routingKey),
			huh.NewSelect[string]().
				Title("Default Severity").
				Options(
					huh.NewOption("Info", "info"),
					huh.NewOption("Warning", "warning"),
					huh.NewOption("Error", "error"),
					huh.NewOption("Critical", "critical"),
				).
				Value(&severity),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["routing_key"] = routingKey
	existing.Options["severity"] = severity

	fmt.Println("  PagerDuty settings updated.")
	return nil
}

// --- Pushbullet ---

func addPushbullet(cfg *config.Config) error {
	var accessToken, deviceIden, channelTag string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Access Token").
				Description("From https://www.pushbullet.com/#settings/account (required)").
				Value(&accessToken),
			huh.NewInput().
				Title("Device Iden").
				Description("Send to specific device only (leave empty for all)").
				Value(&deviceIden),
			huh.NewInput().
				Title("Channel Tag").
				Description("Send to a Pushbullet channel (leave empty for none)").
				Value(&channelTag),
		),
	).Run()
	if err != nil {
		return nil
	}

	options := map[string]interface{}{
		"access_token": accessToken,
	}
	if deviceIden != "" {
		options["device_iden"] = deviceIden
	}
	if channelTag != "" {
		options["channel_tag"] = channelTag
	}

	cfg.Notifiers = append(cfg.Notifiers, config.NotifierConfig{
		Type:    "pushbullet",
		Enabled: true,
		Options: options,
	})

	fmt.Println("  Pushbullet notifier added.")
	return nil
}

func editPushbullet(cfg *config.Config, existing *config.NotifierConfig) error {
	opts := config.WatcherConfig{Options: existing.Options}
	accessToken := opts.GetString("access_token", "")
	deviceIden := opts.GetString("device_iden", "")
	channelTag := opts.GetString("channel_tag", "")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Access Token").
				Value(&accessToken),
			huh.NewInput().
				Title("Device Iden").
				Description("Leave empty for all devices").
				Value(&deviceIden),
			huh.NewInput().
				Title("Channel Tag").
				Description("Leave empty for none").
				Value(&channelTag),
		),
	).Run()
	if err != nil {
		return nil
	}

	existing.Options["access_token"] = accessToken
	if deviceIden != "" {
		existing.Options["device_iden"] = deviceIden
	} else {
		delete(existing.Options, "device_iden")
	}
	if channelTag != "" {
		existing.Options["channel_tag"] = channelTag
	} else {
		delete(existing.Options, "channel_tag")
	}

	fmt.Println("  Pushbullet settings updated.")
	return nil
}
