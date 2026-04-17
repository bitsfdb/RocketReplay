# 🚗 Rocket League Replay Fetcher (Discord Bot)

A lightweight, reverse-engineered Discord bot written in Go that fetches your recent Rocket League matches and provides direct CDN download links for the `.replay` files—without needing to open the game!

This bot uses a multi-step Epic Games authentication flow to communicate directly with Psyonix's servers (PsyNet), pulling match metadata and temporary CDN URLs for un-saved replays.

## ✨ Features
* **Zero Local Storage Bloat:** The bot provides direct Epic CDN links instead of downloading/uploading files locally.
* **Interactive UI:** Uses Discord Modals and Select Menus for a clean, professional user experience.
* **Detailed Match Data:** Calculates match duration, overtime, and sorts player usernames by Team Blue/Orange in the dropdown menu.

## 🛠️ Prerequisites
* [Go 1.24+](https://go.dev/dl/)
* A Discord Bot Token (with `applications.commands` and basic message permissions)
* Epic Games Account

## 🚀 Setup & Installation

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/YOUR_USERNAME/rocket-league-replay-bot.git](https://github.com/YOUR_USERNAME/rocket-league-replay-bot.git)
   cd rocket-league-replay-bot
   ```

2. **Set up your environment variables:**
   Create a `.env` file in the root directory and add your Discord bot token:
   ```env
   DISCORD_BOT_TOKEN=your_token_here
   ```

3. **Run the bot:**
   ```bash
   go run main.go
   ```
   *(We recommend using `screen` or `tmux` to keep the bot running in the background on your server).*

## 🎮 How to Use

Because Epic Games strictly controls third-party API access, this bot uses a trusted Epic Games Launcher Client ID to authenticate. **Users must generate a temporary 32-character exchange code to fetch their replays.**

1. Type `/fetch_replays` in your Discord server.
2. The bot will reply privately with a specific Epic Games login link. Click it and log in.
3. **Important:** After logging in, your browser might redirect to a broken page that says "This site can't be reached" or "localhost refused to connect". **This is completely normal!**
4. Look at the URL bar at the top of your browser. It will look like this:
   `https://localhost/launcher/authorized?code=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`
5. Copy that 32-character code, click the **Enter Code** button in Discord, and paste it in.
6. Select your match from the dropdown menu and grab your replay!

## ⚠️ Disclaimer
This project is for educational purposes only. It is **not** affiliated with, endorsed by, or connected to Epic Games or Psyonix. Use at your own risk. Generating tokens using the Epic Games Launcher Client ID mimics official traffic, but abuse of this API could result in account bans. Do not share your exchange codes with anyone you do not trust.

## 🤝 Credits
Massive shoutout to [dank/rlapi](https://github.com/dank/rlapi) for the Go wrapper around the PsyNet/Epic Games authentication flow.

