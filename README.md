# libdns-gyro

[libdns](https://github.com/libdns/libdns) provider for **Gyro DNS**, plus `gyro-acme-exec`, a tiny static binary
for the lego/Traefik `exec` DNS provider. Both solve ACME **DNS-01** challenges (Let's Encrypt, wildcard certificates
included) with a Gyro DNS **ACME key**.

*Türkçe açıklama aşağıda.*

## ACME key
Create it in the Gyro DNS service page → **ACME / Let's Encrypt** → *Create key* (or `POST /cloud/dns/{id}/acme-keys`,
`gyro dns acme-key create`). An ACME key can only add and remove `_acme-challenge` TXT records of the zones you select;
it cannot read or change any other record, service, invoice or account data. You can also lock it to IP addresses.
If it leaks, delete it in the panel.

## libdns provider (Go)
```go
import gyro "github.com/Gyro-Hosting/libdns-gyro"

p := &gyro.Provider{APIToken: os.Getenv("GYRO_TOKEN")}
// p implements libdns.RecordAppender and libdns.RecordDeleter (TXT _acme-challenge records only)
```
| Field | JSON | Default |
|---|---|---|
| `APIToken` | `api_token` | — (required) |
| `APIBase` | `api_base` | `https://www.gyrohosting.com/api/v1/client/` |
| `SkipWait` | `skip_wait` | `false`: after adding, wait (≤2 min) until both Gyro DNS name servers publish the value |

For **Caddy** use [caddy-dns-gyro](https://github.com/Gyro-Hosting/caddy-dns-gyro).

## gyro-acme-exec (Traefik, lego CLI)
Download from [Releases](https://github.com/Gyro-Hosting/libdns-gyro/releases) (linux/amd64, linux/arm64, darwin, windows).

Traefik (`traefik.yml`):
```yaml
certificatesResolvers:
  le:
    acme:
      email: you@example.com
      storage: /letsencrypt/acme.json
      dnsChallenge:
        provider: exec
```
Environment of the Traefik container:
```
EXEC_PATH=/usr/local/bin/gyro-acme-exec
GYRO_TOKEN=wck_...            # or GYRO_TOKEN_FILE=/run/secrets/gyro_token
EXEC_PROPAGATION_TIMEOUT=300
EXEC_POLLING_INTERVAL=10
```
Mount the binary into the container (e.g. `-v ./gyro-acme-exec:/usr/local/bin/gyro-acme-exec:ro`) and request
`example.com` and `*.example.com` on a router (`tls.domains`).

lego CLI:
```
EXEC_PATH=./gyro-acme-exec GYRO_TOKEN=wck_... lego --dns exec -d example.com -d '*.example.com' --email you@example.com run
```

## Türkçe
Gyro DNS için libdns sağlayıcısı ve Traefik/lego `exec` sağlayıcısı için küçük `gyro-acme-exec` ikilisi. İkisi de
Let's Encrypt DNS-01 doğrulamasını (joker sertifika dahil) Gyro DNS **ACME anahtarıyla** yapar.
ACME anahtarını Gyro DNS hizmet sayfası → **ACME / Let's Encrypt** bölümünden oluşturun. Bu anahtar yalnız seçtiğiniz
zone'larda `_acme-challenge` TXT kaydı ekleyip silebilir; başka kayda, hizmete ya da hesap bilgisine erişemez, isterseniz
IP ile sınırlanır. Ayrıntılı rehber: https://www.gyrohosting.com/ → Bilgi Merkezi → Gyro DNS.

## License
MIT
