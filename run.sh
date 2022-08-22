# Usage: setenv.sh <bot-token> <ngrok url>
export CLESSY_TOKEN=$1
export CLESSY_WEBHOOK=$2/test
export CLESSY_DB_DIR=_data/db
export CLESSY_EMOJI_PATH=_data
export CLESSY_UNSPLASH_FONT=_data/gill.ttf
export CLESSY_MEME_FONT=_data/impact.ttf
export CLESSY_SNAPCHAT_FONT=_data/source.ttf
mkdir -p _data/pics
go run .