# LinkedIn Automation Tool (Go/Rod) 🤖

A sophisticated, stealth-focused browser automation tool built in Go (Golang) using the Rod library.
This project serves as a technical proof-of-concept demonstrating advanced browser automation capabilities, human-like behavior simulation, and clean architecture.

> **⚠️ Educational Purpose Only:** This tool is a Proof of Concept (PoC) for technical evaluation. Automating LinkedIn violates their Terms of Service. Do not use on primary accounts.

## 🎯 Project Objective

> **The primary goal of this project is to build a Go-based linkedin automation tool that showcases:**

- **Advanced Browser Automation:** Using the Rod library.

- **Human-Like Behavior Simulation:** Mimicking authentic user patterns.

- **Sophisticated Anti-Bot Detection:** Implementing robust stealth mechanisms.

## 🚀 Core Features

- **Authentication System** :
  Securely logs in using environment-based credentials, handles failures and security checkpoints, and persists session cookies for seamless reuse.

- **Search & Targeting** : Searches LinkedIn profiles by keywords, handles pagination and infinite scroll, efficiently stores profile URLs, and prevents duplicate processing.

- **Connection Requets** : Programmatically navigates to profiles, accurately triggers connection actions, sends personalized notes, and enforces strict daily limits.

- **Messaging System** : Automatically detects newly accepted connections, sends personalized follow-up messages using templates, and tracks message history to avoid duplicates.

## ✨ Key Features & Capabilities

### 🛡️ Anti-Bot Detection

- **Human-like Mouse Movement:** Simulates organic cursor trajectories using Bézier curves with randomized speed, overshoot, and micro-corrections to bypass detection.
- **Randomized Timing Patterns:** Mimics human cognitive processing by introducing variable "think times" and randomized interaction intervals between all actions.

- **Browser Fingerprint Masking:** Evades fingerprinting by rotating User-Agents on every session and scrubbing automation flags via go-rod/stealth.

### 🥷 Additional Stealth Techniques

- **Random Scrolling Behavior:** Replicates natural reading patterns with inertia-based scrolling, acceleration, and occasional scroll-back movements.

- **Realistic Typing Simulation:** Simulates authentic data entry with variable words-per-minute (WPM), rhythmic hesitations, and automated typo self-correction.

- **Mouse Hovering & Movement:** Enhances realism by injecting random element hovers and natural cursor wandering during idle periods.

- **Activity Scheduling:** Adheres to strict business-hour operation windows and implements heuristic "coffee break" protocols to match human work schedules.

- **Rate Limiting:** Prevents account flagging by enforcing strict daily invite caps and intelligent message throttling.

## 🏗️ Code Architecture & Quality

> **This project adheres to Go best practices and maintains a modular architecture**

- **Modular Design:** Code is organized into logical packages: authentication, search, linkedin (messaging/connect), stealth, guard, and config.

- **Robust Error Handling:** Comprehensive error detection, graceful degradation (e.g., switching to manual login), and retry mechanisms.

- **Structured Logging:** Leveled logging (Info, Warn, Error) provides detailed context for every action.

- **Configuration Management:** Uses .env for all settings (credentials, limits, templates) with sensible defaults.

- **State Persistence:** SQLite database (linkedin.db) tracks all requests and prevents data loss.

## 📂 Directory Structure

> **The project follows a modular, layered architecture aligned with Go best practices.**

```text
├── cmd/
│   └── bot/                 # Application entry point (main.go)
├── internal/
│   ├── browser/             # Rod browser setup & fingerprint config
│   ├── config/              # Environment & config loading
│   ├── guard/               # Rate limits, scheduling, safety rules
│   ├── linkedin/            # Core automation logic
│   │   ├── auth.go
│   │   ├── search.go
│   │   ├── connect.go
│   │   └── message.go
│   ├── stealth/             # Human behavior simulation
│   │   ├── mouse.go
│   │   └── timing.go
│   └── storage/             # SQLite persistence layer
├── .env.example             # Environment template
├── go.mod                   # Go module configuration
└── README.md
```

## ⚙️ Setup & Installation

**Prerequisites**

- Go 1.21 or higher
- Google Chrome installed

**Clone the Repository**

```bash
git clone https://github.com/SNKT2024/linkedin-automation.git
cd linkedin-automation
```

**Install Dependencies**

```bash
go mod tidy
```

**Configure Environment**

```bash
cp .env.example .env
```

_Edit `.env` and add your LinkedIn credentials and configuration preferences._

**Build & Run**

```bash
# Login
go run cmd/bot/main.go --mode=login

# Search for profiles
go run cmd/bot/main.go --mode=search

# Send connection requests
go run cmd/bot/main.go --mode=connect

# Send follow-up messages
go run cmd/bot/main.go --mode=message
```

## 🎥 Demonstration Video

A full walkthrough demonstrating setup, configuration, execution, and core features:

**▶️ [Watch the demo video]()**

The video demonstrates:

- Setup
- Execution
- Stealth behavior
- Logs and safety mechanisms
- Execution of Login/Search/Connect/Message modes, and key stealth features.
