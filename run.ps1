# PowerShell script to run the Flowfull Go Starter
# Usage: .\run.ps1

Write-Host "🚀 Flowfull Go Starter" -ForegroundColor Cyan
Write-Host ""

# Check if .env exists
if (-Not (Test-Path ".env")) {
    Write-Host "⚠️  .env file not found!" -ForegroundColor Yellow
    Write-Host "Creating .env from .env.example..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "✅ .env file created!" -ForegroundColor Green
    Write-Host ""
    Write-Host "⚠️  IMPORTANT: Edit .env file with your configuration before running!" -ForegroundColor Yellow
    Write-Host "Minimum required:" -ForegroundColor Yellow
    Write-Host "  - DATABASE_URL (use sqlite://./flowfull.db for quick start)" -ForegroundColor Yellow
    Write-Host "  - FLOWLESS_API_URL" -ForegroundColor Yellow
    Write-Host "  - BRIDGE_VALIDATION_SECRET (min 32 chars)" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Press Enter to continue after editing .env, or Ctrl+C to exit..."
    Read-Host
}

Write-Host "📦 Installing dependencies..." -ForegroundColor Cyan
go mod download
go mod tidy

Write-Host ""
Write-Host "🏗️  Building application..." -ForegroundColor Cyan
go build -o server.exe ./cmd/server

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful!" -ForegroundColor Green
    Write-Host ""
    Write-Host "🚀 Starting server..." -ForegroundColor Cyan
    Write-Host ""
    .\server.exe
} else {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

