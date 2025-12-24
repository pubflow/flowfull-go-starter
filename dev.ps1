# PowerShell script to run the Flowfull Go Starter in development mode
# Usage: .\dev.ps1

Write-Host "🚀 Flowfull Go Starter - Development Mode" -ForegroundColor Cyan
Write-Host ""

# Check if .env exists
if (-Not (Test-Path ".env")) {
    Write-Host "⚠️  .env file not found!" -ForegroundColor Yellow
    Write-Host "Creating .env from .env.example..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "✅ .env file created!" -ForegroundColor Green
    Write-Host ""
    Write-Host "⚠️  IMPORTANT: Edit .env file with your configuration!" -ForegroundColor Yellow
    Write-Host "For quick start, use SQLite:" -ForegroundColor Yellow
    Write-Host "  DATABASE_URL=sqlite://./flowfull.db" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Press Enter to continue after editing .env, or Ctrl+C to exit..."
    Read-Host
}

Write-Host "📦 Installing dependencies..." -ForegroundColor Cyan
go mod download

Write-Host ""
Write-Host "🔥 Running in development mode (with auto-reload)..." -ForegroundColor Cyan
Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
Write-Host ""

# Run with go run (will reload on file changes if using air, otherwise just run)
go run ./cmd/server/main.go

