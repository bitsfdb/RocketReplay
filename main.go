package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dank/rlapi"
	"github.com/joho/godotenv"
)

var (
	BotToken   string
	matchCache sync.Map
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, relying on system environment variables.")
	}
	BotToken = os.Getenv("DISCORD_BOT_TOKEN")
	if BotToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN is not set in the environment.")
	}
}

func main() {
	dg, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// 1. Updated Router: Handles Commands, Buttons, Modals, and Dropdowns!
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == "fetch_replays" {
				handleFetchCommand(s, i)
			}
		case discordgo.InteractionMessageComponent:
			switch i.MessageComponentData().CustomID {
			case "btn_enter_code":
				handleEnterCodeButton(s, i)
			case "select_replay":
				handleReplaySelection(s, i)
			}
		case discordgo.InteractionModalSubmit:
			if i.ModalSubmitData().CustomID == "modal_enter_code" {
				handleCodeSubmit(s, i)
			}
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	// 2. We removed the options! It's just a simple /fetch_replays now
	command := &discordgo.ApplicationCommand{
		Name:        "fetch_replays",
		Description: "Fetches your last 20 Rocket League matches securely.",
	}

	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", command)
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// ==========================================
// STEP 1: SHOW THE LINK AND BUTTON
// ==========================================
func handleFetchCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	instructions := "**Step 1:** Click the link below and log into Epic Games.\n" +
		"**Step 2:** Wait for the page to break/error out.\n" +
		"**Step 3:** Look at the URL bar and copy the 32-character code at the end.\n" +
		"**Step 4:** Click the 'Enter Code' button below!\n\n" +
		"🔗 **[Click Here to Log In](https://www.epicgames.com/id/login?redirectUrl=https%3A%2F%2Fwww.epicgames.com%2Fid%2Fapi%2Fredirect%3FclientId%3D34a02cf8f4414e29b15921876da36f9a%26responseType%3Dcode)**"

	button := discordgo.Button{
		Label:    "Enter Code",
		Style:    discordgo.PrimaryButton,
		CustomID: "btn_enter_code",
		Emoji: &discordgo.ComponentEmoji{
			Name: "🔑",
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: instructions,
			Flags:   discordgo.MessageFlagsEphemeral, // Keeps it private
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{button},
				},
			},
		},
	})
}

// ==========================================
// STEP 2: OPEN THE POP-UP MODAL
// ==========================================
func handleEnterCodeButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modal_enter_code",
			Title:    "Epic Games Authentication",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "exchange_code_input",
							Label:       "32-Character Exchange Code",
							Style:       discordgo.TextInputShort,
							Placeholder: "Paste code here...",
							Required:    true,
							MinLength:   32,
							MaxLength:   32,
						},
					},
				},
			},
		},
	})
}

// ==========================================
// STEP 3: AUTHENTICATE & SHOW DROPDOWN MENU
// ==========================================
func handleCodeSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge the modal submission immediately
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	userID := getUserID(i)
	
	// Extract the code from the Modal
	exchangeCode := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	// Let the user know we are working on it
	loadingMsg := "🔄 Authenticating with Epic Games and connecting to PsyNet..."
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &loadingMsg,
		Components: &[]discordgo.MessageComponent{}, // Clear the button
	})

	egsClient := rlapi.NewEGS()
	tokenResp, err := egsClient.AuthenticateWithCode(exchangeCode)
	if err != nil {
		sendError(s, i, "Failed to authenticate with Epic. Code might be expired.")
		return
	}

	eosExchangeCode, err := egsClient.GetExchangeCode(tokenResp.AccessToken)
	if err != nil {
		sendError(s, i, "Failed to generate EOS exchange code.")
		return
	}

	eosTokenResp, err := egsClient.ExchangeEOSToken(eosExchangeCode)
	if err != nil {
		sendError(s, i, "Failed to get EOS token.")
		return
	}

	psyNetClient := rlapi.NewPsyNet()
	rpcClient, err := psyNetClient.AuthPlayer(eosTokenResp.AccessToken, tokenResp.AccountID, tokenResp.DisplayName)
	if err != nil {
		sendError(s, i, "Failed to connect to Rocket League PsyNet.")
		return
	}

	history, err := rpcClient.GetMatchHistory(context.Background())
	if err != nil {
		sendError(s, i, "Failed to fetch match history from Psyonix servers.")
		return
	}

	matchCache.Store(userID, history)

	var menuOptions []discordgo.SelectMenuOption
	validMatches := 0

	for idx, matchEntry := range history {
		if matchEntry.ReplayUrl == "" {
			continue
		}

		match := matchEntry.Match
		validMatches++

		matchTime := time.Unix(match.RecordStartTimestamp, 0).Format("Jan 02, 15:04")
		totalSeconds := int(match.SecondsPlayed + match.OvertimeSecondsPlayed)
		duration := fmt.Sprintf("%dm %ds", totalSeconds/60, totalSeconds%60)
		if match.OverTime {
			duration += " (OT)"
		}

		var blueNames, orangeNames []string
		for _, p := range match.Players {
			if p.LastTeam == 0 {
				blueNames = append(blueNames, p.PlayerName)
			} else {
				orangeNames = append(orangeNames, p.PlayerName)
			}
		}

		label := fmt.Sprintf("🔵 %d - %d 🟠 | %s", match.Team0Score, match.Team1Score, matchTime)
		desc := fmt.Sprintf("B: %s | O: %s (%s)", strings.Join(blueNames, ", "), strings.Join(orangeNames, ", "), duration)
		if len(desc) > 95 {
			desc = desc[:95] + "..."
		}

		menuOptions = append(menuOptions, discordgo.SelectMenuOption{
			Label:       label,
			Description: desc,
			Value:       strconv.Itoa(idx),
		})

		if len(menuOptions) >= 25 {
			break
		}
	}

	if validMatches == 0 {
		sendError(s, i, "Found matches, but none of them have replays available to download!")
		return
	}

	selectMenu := discordgo.SelectMenu{
		CustomID:    "select_replay",
		Placeholder: "Choose a match to download...",
		Options:     menuOptions,
	}

	actionRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{selectMenu},
	}

	successMsg := fmt.Sprintf("✅ Auth successful! Found %d available replays. Pick one below:", validMatches)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &successMsg,
		Components: &[]discordgo.MessageComponent{actionRow},
	})
}

// ==========================================
// STEP 4: PROVIDE CDN LINK
// ==========================================
func handleReplaySelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	userID := getUserID(i)

	cacheData, ok := matchCache.Load(userID)
	if !ok {
		sendFollowupError(s, i, "Session expired! Please run /fetch_replays again.")
		return
	}

	history := cacheData.([]rlapi.MatchEntry)
	selection := i.MessageComponentData().Values[0]
	idx, err := strconv.Atoi(selection)
	if err != nil || idx >= len(history) {
		sendFollowupError(s, i, "Invalid selection.")
		return
	}

	matchEntry := history[idx]

	responseMsg := fmt.Sprintf("✅ **Here is your direct CDN link!**\n\n**Match GUID:** `%s`\n**Download Link:** %s\n\n*(Note: Epic CDN links expire quickly, click it soon!)*",
		matchEntry.Match.MatchGUID,
		matchEntry.ReplayUrl,
	)

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: responseMsg,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}

// ==========================================
// HELPERS
// ==========================================
func getUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.ID
	}
	return i.User.ID
}

func sendError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	formattedMsg := "❌ " + msg
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &formattedMsg,
	})
}

func sendFollowupError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	formattedMsg := "❌ " + msg
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: formattedMsg,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
}
