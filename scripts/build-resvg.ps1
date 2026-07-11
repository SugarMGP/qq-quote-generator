$ErrorActionPreference = "Stop"

$Version = "0.47.0"
$Target = "x86_64-pc-windows-gnu"
$Source = Join-Path $env:TEMP "qq-quote-resvg-$Version"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$Output = Join-Path $ProjectRoot "internal/resvg/lib/windows-amd64"

if (-not (Get-Command git -ErrorAction SilentlyContinue)) { throw "git is required" }
if (-not (Get-Command cargo -ErrorAction SilentlyContinue)) { throw "cargo is required" }
if (-not (Get-Command rustup -ErrorAction SilentlyContinue)) { throw "rustup is required" }
$Gcc = Get-Command gcc -ErrorAction SilentlyContinue
if (-not $Gcc) { throw "MinGW-w64 gcc is required" }
$env:CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER = $Gcc.Source

if (-not (Test-Path $Source)) {
    git clone --depth 1 --branch "v$Version" https://github.com/linebender/resvg $Source
    if ($LASTEXITCODE -ne 0) { throw "failed to clone resvg v$Version" }
}

$Commit = git -C $Source describe --tags --exact-match
if ($LASTEXITCODE -ne 0) { throw "failed to inspect cached resvg version" }
if ($Commit -ne "v$Version") { throw "resvg cache is not v${Version}: $Commit" }

rustup target add $Target
if ($LASTEXITCODE -ne 0) { throw "failed to install Rust target $Target" }
cargo build --release -p resvg-capi --target $Target --manifest-path (Join-Path $Source "Cargo.toml")
if ($LASTEXITCODE -ne 0) { throw "failed to build resvg $Version" }

New-Item -ItemType Directory -Force $Output | Out-Null
$Library = Join-Path $Source "target/$Target/release/libresvg.a"
if (-not (Test-Path $Library)) { throw "resvg static library was not produced: $Library" }
Copy-Item $Library (Join-Path $Output "libresvg.a") -Force
Copy-Item (Join-Path $Source "crates/c-api/resvg.h") (Join-Path $ProjectRoot "internal/resvg/resvg.h") -Force

Write-Host "resvg $Version built at $Output"
