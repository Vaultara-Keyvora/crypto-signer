# Crypto Signer

A self-hosted signing service for multi-chain HD wallets (Ethereum, BSC, TRON, Solana).
It derives addresses from BIP39/BIP32 mnemonics held in process memory and signs
transactions requested by a merchant platform over HTTP. Requests are authenticated
with an HMAC-SHA256 shared secret. A small embedded web page is included for local
configuration.

> Run it only on a trusted host. Never commit real configuration, keys or mnemonics.

## Requirements

- Go 1.25+ (to build from source), or
- Docker

## Build

```bash
go build -o bin/crypto_signer ./cmd/active
# or
make build
```

## Configure

Configuration lives in `config/`:

| File | Purpose |
|------|---------|
| `api.yaml` | listen address, port (default `10013`), trusted proxies, enabled chains |
| `log.yaml` | log location, rotation and level |
| `signer.example.yaml` | template for `signer.yaml` (merchant/signer settings) |

`config/signer.yaml` holds secrets and is git-ignored. It must exist before the
service or `init` starts, because both load it at startup:

```bash
cp config/signer.example.yaml config/signer.yaml
```

Then either edit the values in `config/signer.yaml` by hand, or fill them in with
`init` (run from the repository root):

```bash
./bin/crypto_signer init \
  --account      <YOUR_ETH_ADDRESS> \
  --merchant-id  <YOUR_MERCHANT_ID> \
  --merchant-url https://<YOUR_MERCHANT_API_HOST>/api/ \
  --signer-url   http://<YOUR_SIGNER_HOST>:10013 \
  --secret-key   <YOUR_SECRET_KEY>
```

`init` overwrites `./config/signer.yaml` (relative to the current directory) and then
registers the signer with the merchant platform, so the merchant URL must be
reachable. `--secret-key` must be at least 32 characters.

## Run

```bash
./bin/crypto_signer -c ./config
```

## Run with Docker

The image is distroless, has no shell and runs as a non-root user (UID 65532).
Configuration is not baked into the image; mount your `config/` directory.

```bash
docker build -t crypto-signer .

# Inside a container the service must listen on all interfaces:
# set `listenIP: 0.0.0.0` in config/api.yaml first.
docker run -d --name crypto-signer \
  -p 10013:10013 \
  -v "$PWD/config:/app/config:ro" \
  crypto-signer
```

Create `config/signer.yaml` before starting the container (see **Configure**).
