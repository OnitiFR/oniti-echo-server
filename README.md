# Oniti Echo Server

SSE event server for Laravel "vlank" applications

## Download
`go install github.com/OnitiFR/oniti-echo-server@latest`

## Run (dev)
`go install && oniti-echo-server`


## Build & deploy (Oniti only)
`CGO_ENABLED=0 go build && scp oniti-echo-server files-m4:public_html/data/oniti/`
