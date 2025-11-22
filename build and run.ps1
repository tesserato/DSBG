go get -u
go get -u ./...
go mod tidy

gofmt -s -w .

go vet ./...

# Kill any existing instances of dsbg to free up file locks and ports
Stop-Process -Name "dsbg" -ErrorAction SilentlyContinue

Remove-Item ./dsbg.exe -Force -ErrorAction SilentlyContinue
go build .

./dsbg.exe -h

# Remove-Item "docs/*" -Recurse -Force

Copy-Item "README.md" "sample_content/01_readme.md" -Force
magick "art/logo.webp" "sample_content/01_dsbg_logo.webp"
magick -background none "sample_content/01_dsbg_logo.webp" -fill red -opaque black -blur 0x1  -crop 167x167+0+0  "src/assets/favicon.ico"
# magick -background none "sample_content/01_dsbg_logo.webp"  -crop 167x167+0+0  "thumb.webp"

$description = @'
A Simple, Open-Source Tool to Create Your Static Blog and Broadcast Your Content.

# TLDR:

`go install github.com/tesserato/DSBG@latest` or download a [pre-built binary](https://github.com/tesserato/DSBG/releases)

`dsbg -h` for usage instructions

Check the Readme [here](https://tesserato.github.io/DSBG/01readme/index.html) or at the Github [repo](https://github.com/tesserato/DSBG) for more details

This is a sample blog created with DSBG from the source at [github.com/tesserato/DSBG](https://github.com/tesserato/DSBG/tree/main/sample_content)

[![Release Status](https://img.shields.io/github/release/tesserato/DSBG)](https://github.com/tesserato/DSBG/releases)

[![License](https://img.shields.io/github/license/tesserato/DSBG)](https://github.com/tesserato/DSBG/blob/main/LICENSE)

'@

# [![Build Status](https://github.com/tesserato/DSBG/actions/workflows/go.yml/badge.svg)](https://github.com/tesserato/DSBG/actions/workflows/go.yml)
# [![Go Version](https://img.shields.io/github/go-mod/go-version/tesserato/DSBG)](https://go.dev/)
# Start-Process chrome http://localhost:666/index.html


# $theme = "default"
# $theme = "dark"
# $theme = "black"
$theme = "industrial"

# Base URL for the raw assets on GitHub
$ghAssets = "https://raw.githubusercontent.com/tesserato/DSBG/refs/heads/main/sample_assets"

./dsbg.exe -title "Dead Simple Blog Generator" `
    -description "$description" `
    -overwrite `
    -watch `
    -input "sample_content" `
    -output "docs" `
    -base-url "https://tesserato.github.io/DSBG/" `
    -elements-top "sample_assets/analytics.html" `
    -elements-bottom "sample_assets/giscus.html" `
    -theme $theme `
    -share "X|$ghAssets/x.svg|https://x.com/intent/tweet?text={DESCRIPTION}&url={URL}" `
    -share "Telegram|$ghAssets/telegram.svg|https://t.me/share/url?url={URL}&text={DESCRIPTION}" `
    -share "LinkedIn|$ghAssets/linkedin.svg|https://www.linkedin.com/sharing/share-offsite/?url={URL}" `
    -share "Reddit|$ghAssets/reddit.svg|https://www.reddit.com/submit?url={URL}&title={TITLE}" `
    -share "Bluesky|$ghAssets/bluesky.svg|https://bsky.app/intent/compose?text={DESCRIPTION}%20{URL}" `
    -share "Mastodon|$ghAssets/mastodon.svg|https://mastodon.social/share?text={DESCRIPTION}%20{URL}" `
    -share "Threads|$ghAssets/threads.svg|https://www.threads.net/intent/post?text={DESCRIPTION}%20{URL}" `
    -share "Hacker News|$ghAssets/hackernews.svg|https://news.ycombinator.com/submitlink?u={TARGET_URL}&t={TITLE}" `
    -share "Facebook|$ghAssets/facebook.svg|https://www.facebook.com/sharer/sharer.php?u={URL}" `
    -sort "reverse-date-created" `
    -css-path "./src/assets/themes/default.css"    
    