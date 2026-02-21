# AgentPay Setup Guide

## Installing Go

You have two options to install Go on macOS:

### Option 1: Install via Homebrew (Recommended)

First, fix Homebrew permissions:
```bash
sudo chown -R $(whoami) /opt/homebrew
sudo chown -R $(whoami) /Users/$(whoami)/Library/Logs/Homebrew
```

Then install Go:
```bash
brew install go
```

### Option 2: Install via Official Go Installer

1. Download Go from: https://go.dev/dl/
2. Download the macOS installer (e.g., `go1.21.x.darwin-amd64.pkg`)
3. Run the installer and follow the prompts
4. Verify installation:
   ```bash
   go version
   ```

### Verify Go Installation

After installation, verify it works:
```bash
go version
```

You should see something like: `go version go1.21.x darwin/arm64`

## Setting Up the Project

Once Go is installed:

1. **Start the database:**
   ```bash
   docker-compose up -d
   ```

2. **Install Go dependencies:**
   ```bash
   cd backend
   go mod download
   ```

3. **Run the backend:**
   ```bash
   export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=agentpay PORT=8080
   go run cmd/api/main.go
   ```

4. **In another terminal, set up and run the frontend:**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Troubleshooting

- If `go` command is still not found after installation, you may need to restart your terminal or add Go to your PATH:
  ```bash
  export PATH=$PATH:/usr/local/go/bin
  ```
  Add this to your `~/.zshrc` to make it permanent.
