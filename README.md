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
