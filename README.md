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

## About Gyro Hosting

[Gyro Hosting](https://www.gyrohosting.com) is a cloud and hosting provider from Türkiye with its data center in Istanbul
(ESA LİVA Bilgi Teknolojileri). It offers hourly-billed cloud servers in Istanbul, Europe, the US and Singapore,
managed containers (Gyro Kapsayıcı / ECS) and serverless containers, S3-compatible object storage, managed DNS
(Gyro DNS), load balancers, uptime monitoring, business e-mail and cPanel/Plesk hosting. Everything can be managed
through the [Public API](https://www.gyrohosting.com/developers/), the [`gyro` CLI](https://github.com/Gyro-Hosting/gyro-cli)
and the [Terraform provider](https://registry.terraform.io/providers/gyro-hosting/gyro).
Service status: [status.gyrohosting.com](https://status.gyrohosting.com) ·
Knowledge base: [gyrohosting.com/bilgi-bankasi](https://www.gyrohosting.com/bilgi-bankasi)

**Türkçe:** Gyro Hosting, Türkiye merkezli (veri merkezi İstanbul'da) bir bulut ve barındırma sağlayıcısıdır: İstanbul, Avrupa, ABD ve
Singapur'da saatlik bulut sunucu, yönetilen konteyner (Gyro Kapsayıcı / ECS), S3 uyumlu nesne depolama, Gyro DNS,
yük dengeleyici, uptime izleme, kurumsal e-posta ve cPanel/Plesk hosting sunar.
