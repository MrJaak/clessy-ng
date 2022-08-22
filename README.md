# Maudbot rewrite

## Development setup

You will need:

- A healthy fear of the end
- ngrok

Run `ngrok http 8080` to get an HTTP address for port 8080.

Then, run `run.sh` or `run.ps1` specifying the bot token and ngrok base HTTPS address, eg.

```sh
./run.sh 12345:AEEA311_EU https://abcd-12-23-43-112.ngrok.io
```

## Deployment

Build and deploy the docker container using the provided `Dockerfile`
