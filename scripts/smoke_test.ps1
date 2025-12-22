param(
  [string]$DbName = "amss_smoke",
  [string]$BaseUrl = "http://127.0.0.1:8080",
  [switch]$KeepRunning
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
  throw "docker not found in PATH"
}
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
  throw "go not found in PATH"
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Push-Location $repoRoot
$serverProcess = $null

try {
  & docker compose up -d postgres redis | Out-Null

  $ready = $false
  for ($i = 0; $i -lt 30; $i++) {
    & docker compose exec -T postgres pg_isready -U amss | Out-Null
    if ($LASTEXITCODE -eq 0) {
      $ready = $true
      break
    }
    Start-Sleep -Seconds 1
  }
  if (-not $ready) {
    throw "postgres did not become ready"
  }

  $createDbSql = "DROP DATABASE IF EXISTS $DbName; CREATE DATABASE $DbName;"
  $createDbSql | docker compose exec -T postgres psql -U amss -d postgres -v ON_ERROR_STOP=1 | Out-Null

  $files = Get-ChildItem -Path "migrations" -Filter "*.sql" | Sort-Object Name
  foreach ($file in $files) {
    $sql = Get-Content $file.FullName -Raw
    $idx = $sql.IndexOf("-- +goose Down")
    if ($idx -ge 0) {
      $sql = $sql.Substring(0, $idx)
    }
    $sql = $sql.Trim()
    if ($sql.Length -eq 0) {
      continue
    }
    $sql | docker compose exec -T postgres psql -U amss -d $DbName -v ON_ERROR_STOP=1 | Out-Null
  }

  $importDir = Join-Path (Join-Path $repoRoot "data") "imports"
  New-Item -ItemType Directory -Force -Path $importDir | Out-Null

  $privatePem = $null
  $publicPem = $null
  try {
    $rsa = [System.Security.Cryptography.RSA]::Create(2048)
    if ($rsa -and $rsa.PSObject.Methods.Match("ExportPkcs8PrivateKey").Count -gt 0 -and $rsa.PSObject.Methods.Match("ExportSubjectPublicKeyInfo").Count -gt 0) {
      $privateKey = $rsa.ExportPkcs8PrivateKey()
      $publicKey = $rsa.ExportSubjectPublicKeyInfo()
      $privatePem = "-----BEGIN PRIVATE KEY-----`n" + [Convert]::ToBase64String($privateKey, "InsertLineBreaks") + "`n-----END PRIVATE KEY-----"
      $publicPem = "-----BEGIN PUBLIC KEY-----`n" + [Convert]::ToBase64String($publicKey, "InsertLineBreaks") + "`n-----END PUBLIC KEY-----"
    }
  } catch {
  }
  if (-not $privatePem -or -not $publicPem) {
    $privatePem = @'
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAubxcEOrP2b7gm0uPUL3lkGWRavBelIF3JLEKnytCIUBNIMYg
+WqJYvdN8jhhQWSKcGehKDUEcsk1+mwStFZhxmbGq7hqfJTWbhd5KVa28h3U0MRn
jbWzwrgX0lHatZQFo29Lc4tgz23mNs2M2uEfYWqc39oulmN3wKT/XvSs+b4DHPZN
bojD+gG1/cuA+OR5/ektVUG94XcfTzJw25MfXR70vCXSOQLfFsxUTUY2ObzEdiEw
5XUPCD4LlYa2OdDJotxn6Dz7YhVcYmrUNs2V+/PyOKrxnlyckNIgIHnfBfbWYJqp
0YhZhe8FkDnQwZJkWzrkqluRvDomI5x2BQL6qQIDAQABAoIBAAzLW2I08qVyvyEj
dtdehWpJfNdrgHBlbLIj2fH19mO4/Lclvs6/loAxPfbSCG0LQngUw2a0JD7N+oBq
Im22x5x1pvjcRsWXMSA7ULbWyNpr0KWi6ek6m4vtx4JO7ix06mCEQaPPdZdEYEfl
J+9J87HvuKe47V2hs3WbSBYf76xyWtOiCobjNFIK2lkEch5/ezfL+H557nRYiO/2
FSMWmzrzSbFAyzsvkpQBgZd89/WAU4ngN8stt0MLbRTHMETWQg6c6rrF+LCxRYbF
uDeIniLbgusQQZ/vRTKQ+MOBApK3vUyTJbFAUvplwQYvJOGPlmSC1JBzURyx6/BG
dfSt25MCgYEA1kOHvx4YFycB4L3jUv7QSh70miEeIKJonPomBfTXKyvxfxsVcu7j
iwN/EmRhZexdXaIVcsLkB4rjduqNXvIIMn8xsAvDUK4Nu2+poNmjRovMXEs2uzoO
Ew7Nffb0KbGINBIU64G+cgfYJCwSTyqywwlIkBhasLuDS54bCdALr4cCgYEA3epA
sy1h8rNY8/agPXX55DWI9WS2USjFkaSibZ0PH665WHOWHVdnGnxfn+3Lo56Wi9FP
NqQ5gbR9a2cNlRaz7U/JJmYsvzi5cmPXWFJtx/CruUeUGVjxckvE4IMS/MREcZQM
F/EIIFPxizrOVQG1stYze0GhEf8KHnkB7yItsE8CgYEAgKLWcsVsjSncFMOsIP3e
q0FedNKBNfKLgAMmpNjT/ZVKTZdDD1egwKr+tVoSp5B6lWZkHhwnrueRnKlA6snA
ZiC7Aght4Jg+olNtsaY4QnhX3ulBGLLIFGUEtiV3fTianzhj2uhwICHZgA39iA4I
eNOv/uLAP+6z6sgnT4LaIS0CgYBDbJEL35YK74LvXNeC1P1/4OQj6t2Z+xFMFwFi
3H1j2uplfXj2oT+qRG+pX86nf9+ty4KNz4fJaNVSdJUj3yn7yGoNSK3/y3RM1Rjw
tNq2DOGgAad1rBhv6aV/sVNriRZii+DAxXL6n4acDtnx6fsSwxIROPd/SEYCzDFS
Psgy8QKBgHcJ4wv/fRYSlIBwCHLPJEQK8kOjB8Px1sdz/9k2E6qqSBRjY3kK8+EB
D2V/soGeb6Wa78ERp4dFzXR2tTiIfVoRmCRh7KWmVM//dcl2Fk7EoC32M7q2B0oj
VwERjL025eugnwKseh6ZdRC2lOWKixiFlMwIMi4JGkiizHCWholS
-----END RSA PRIVATE KEY-----
'@
    $publicPem = @'
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAubxcEOrP2b7gm0uPUL3l
kGWRavBelIF3JLEKnytCIUBNIMYg+WqJYvdN8jhhQWSKcGehKDUEcsk1+mwStFZh
xmbGq7hqfJTWbhd5KVa28h3U0MRnjbWzwrgX0lHatZQFo29Lc4tgz23mNs2M2uEf
YWqc39oulmN3wKT/XvSs+b4DHPZNbojD+gG1/cuA+OR5/ektVUG94XcfTzJw25Mf
XR70vCXSOQLfFsxUTUY2ObzEdiEw5XUPCD4LlYa2OdDJotxn6Dz7YhVcYmrUNs2V
+/PyOKrxnlyckNIgIHnfBfbWYJqp0YhZhe8FkDnQwZJkWzrkqluRvDomI5x2BQL6
qQIDAQAB
-----END PUBLIC KEY-----
'@
  }

  $env:DB_URL = "postgres://amss:amss@127.0.0.1:5455/${DbName}?sslmode=disable"
  $env:REDIS_ADDR = "127.0.0.1:6379"
  $httpPort = $null
  $baseUri = $null
  try {
    $baseUri = [System.Uri]$BaseUrl
  } catch {
  }
  if (-not $baseUri -or -not $baseUri.IsAbsoluteUri) {
    throw "invalid BaseUrl: $BaseUrl"
  }
  if ($BaseUrl -match "127\.0\.0\.1:8080" -or $BaseUrl -match "localhost:8080") {
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
    $listener.Start()
    $httpPort = $listener.LocalEndpoint.Port
    $listener.Stop()
    $BaseUrl = "http://127.0.0.1:$httpPort"
  } else {
    $httpPort = $baseUri.Port
  }

  $env:HTTP_ADDR = ":$httpPort"
  $env:GRPC_ADDR = ":0"
  $env:JWT_PRIVATE_KEY_PEM = $privatePem
  $env:JWT_PUBLIC_KEY_PEM = $publicPem
  $env:IMPORT_STORAGE_DIR = (Resolve-Path $importDir).Path
  $env:APP_ENV = "development"
  $env:PROMETHEUS_ENABLED = "false"
  $env:LOG_LEVEL = "info"

  $serverProcess = Start-Process -FilePath "go" -ArgumentList @("run", "./cmd/server") -WorkingDirectory $repoRoot -NoNewWindow -PassThru

  $serverReady = $false
  $lastReadyError = $null
  for ($i = 0; $i -lt 30; $i++) {
    try {
      $resp = Invoke-WebRequest -Uri "$BaseUrl/ready" -TimeoutSec 2 -UseBasicParsing
      if ($resp.StatusCode -eq 200) {
        $serverReady = $true
        break
      }
      if ($resp.Content) {
        $lastReadyError = $resp.Content.Trim()
      }
    } catch {
      $lastReadyError = $null
      if ($_.Exception -and $_.Exception.Response) {
        try {
          $respStream = $_.Exception.Response.GetResponseStream()
          $reader = New-Object System.IO.StreamReader($respStream)
          $lastReadyError = $reader.ReadToEnd()
          $reader.Close()
        } catch {
        }
      }
    }
    Start-Sleep -Seconds 1
  }
  if (-not $serverReady) {
    if ($lastReadyError) {
      throw "server did not become ready: $lastReadyError"
    }
    throw "server did not become ready"
  }

  $health = Invoke-WebRequest -Uri "$BaseUrl/health" -TimeoutSec 5 -UseBasicParsing
  if ($health.StatusCode -ne 200 -or $health.Content.Trim() -ne "ok") {
    throw "health check failed"
  }

  $orgId = (docker compose exec -T postgres psql -U amss -d $DbName -t -A -c "SELECT id FROM organizations WHERE name='Demo Airline' LIMIT 1;").Trim()
  if ([string]::IsNullOrWhiteSpace($orgId)) {
    throw "seed org not found"
  }

  $loginBody = @{ org_id = $orgId; email = "admin@demo.local"; password = "ChangeMe123!" } | ConvertTo-Json
  $login = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/auth/login" -ContentType "application/json" -Body $loginBody
  if (-not $login.access_token) {
    throw "login failed"
  }

  $headers = @{ Authorization = "Bearer $($login.access_token)" }
  $orgs = Invoke-RestMethod -Uri "$BaseUrl/api/v1/organizations" -Headers $headers
  if ($null -eq $orgs) {
    throw "organizations response empty"
  }

  Write-Host "Smoke test complete"
} finally {
  if (-not $KeepRunning) {
    if ($serverProcess -and -not $serverProcess.HasExited) {
      Stop-Process -Id $serverProcess.Id -Force
    }
    if ($httpPort) {
      try {
        $conn = Get-NetTCPConnection -LocalPort $httpPort -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($conn) {
          Stop-Process -Id $conn.OwningProcess -Force
        }
      } catch {
      }
    }
    Start-Sleep -Seconds 1
  }

  if (-not $KeepRunning) {
    try {
      if ($DbName) {
        "DROP DATABASE IF EXISTS $DbName;" | docker compose exec -T postgres psql -U amss -d postgres -v ON_ERROR_STOP=1 | Out-Null
      }
    } catch {
    }
  }

  if (-not $KeepRunning) {
    try { & docker compose down | Out-Null } catch {}
  }
  Pop-Location
}
