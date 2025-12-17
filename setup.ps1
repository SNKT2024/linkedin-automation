# Create directories
New-Item -ItemType Directory -Path cmd/bot -Force
New-Item -ItemType Directory -Path config -Force
New-Item -ItemType Directory -Path internal/browser -Force
New-Item -ItemType Directory -Path internal/linkedin -Force
New-Item -ItemType Directory -Path internal/stealth -Force
New-Item -ItemType Directory -Path internal/storage -Force

# Create files
New-Item -ItemType File -Path cmd/bot/main.go -Force
New-Item -ItemType File -Path config/config.go -Force
New-Item -ItemType File -Path internal/browser/engine.go -Force
New-Item -ItemType File -Path internal/linkedin/auth.go -Force
New-Item -ItemType File -Path internal/stealth/stealth.go -Force
New-Item -ItemType File -Path internal/storage/sqlite.go -Force
New-Item -ItemType File -Path .env -Force

Write-Host "Folder structure and files created successfully."