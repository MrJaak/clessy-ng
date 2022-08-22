# Usage: setenv.ps1 <bot-token> <ngrok url>
$env:CLESSY_TOKEN = "5763477340:AAFYInOfYsrsHdP64farFJcLyks3_0WuIkw"
$env:CLESSY_WEBHOOK = $args[0] + "/test"
$env:CLESSY_DB_DIR = "_data/db"
$env:CLESSY_EMOJI_PATH = "_data"
$env:CLESSY_UNSPLASH_FONT = "_data/gill.ttf"
$env:CLESSY_UNSPLASH_BG_PATH = "_data/pics"
$env:CLESSY_MEME_FONT = "_data/impact.ttf"
$env:CLESSY_SNAPCHAT_FONT = "_data/source.ttf"
mkdir -force _data/pics
go run .