Set-Location src
go get -u
go get -u ./...
go mod tidy

Remove-Item ./dsbg.exe -Force -ErrorAction SilentlyContinue
go build .

Remove-Item "../docs/*" -Recurse -Force

Copy-Item "../README.md" "../sample_content/01_readme.md" -Force
magick "../art/logo.webp" "../sample_content/01_dsbg_logo.webp"
magick -background none "../sample_content/01_dsbg_logo.webp" -fill red -opaque black -blur 0x1  -crop 167x167+0+0  "assets/favicon.ico"
magick -background none "../sample_content/01_dsbg_logo.webp"  -crop 167x167+0+0  "thumb.webp"

$description = @'
A Simple, Open-Source Tool to Create Your Static Blog and Broadcast Your Content.

# TLDR:

`go install github.com/tesserato/DSBG@latest` or download a [pre-built binary](https://github.com/tesserato/DSBG/releases)

`dsbg -h` for usage instructions

Check the Readme [here](https://tesserato.github.io/dsbg/01readme/index.html) or at the Github [repo](https://github.com/tesserato/DSBG) for more details

This is a sample blog created with DSBG from the source at [github.com/tesserato/DSBG](https://github.com/tesserato/DSBG/tree/main/sample_content)

[![Release Status](https://img.shields.io/github/release/tesserato/DSBG)](https://github.com/tesserato/DSBG/releases)

[![License](https://img.shields.io/github/license/tesserato/DSBG)](https://github.com/tesserato/DSBG/blob/main/LICENSE)

'@

# [![Build Status](https://github.com/tesserato/DSBG/actions/workflows/go.yml/badge.svg)](https://github.com/tesserato/DSBG/actions/workflows/go.yml)
# [![Go Version](https://img.shields.io/github/go-mod/go-version/tesserato/DSBG)](https://go.dev/)

Start-Process chrome http://localhost:666/index.html



./dsbg.exe -title "Dead Simple Blog Generator" `
    -description "$description" `
    -watch `
    -input-path "../sample_content" `
    -output-path "../docs" `
    -base-url "https://tesserato.github.io/DSBG/" `
    -elements-top "../sample_content/analytics.html" `
    -elements-bottom "../sample_content/giscus.html" `
    -theme "dark" `
    -x-handle "tesserato" `
    -bluesky-handle "tesserato" `
    -threads-handle "tesserato" `
    -mastodon-handle "tesserato" `
    -telegram-handle "tesserato" `
    -linkedin-handle "tesserato" `
    -reddit-handle "tarjano" `
    -hackernews-handle "tesserato" `
    -sort "reverse-date-created"
    
    