#!/usr/bin/env bash
# Builds the app icon and favicons in static/ from mark.svg.
# Needs rsvg-convert and ImageMagick (brew install librsvg imagemagick).
# Run from anywhere: misc/icon/build.sh
set -euo pipefail
here="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
out="$here/../../static"
work="$( mktemp -d )"
trap 'rm -rf "$work"' EXIT

mark="$( sed 1d "$here/mark.svg" )"
grad='<defs><linearGradient id="bg" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#2e74ec"/><stop offset="1" stop-color="#1642a0"/></linearGradient></defs>'
head='<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">'

# Rounded tile: the browser-tab icon and the "any" manifest icons.
printf '%s%s<rect width="512" height="512" rx="112" fill="url(#bg)"/>%s</svg>\n' "$head" "$grad" "$mark" > "$out/icon.svg"
# Full bleed for iOS, which rounds the corners itself.
printf '%s%s<rect width="512" height="512" fill="url(#bg)"/><g transform="translate(46 46) scale(0.82)">%s</g></svg>\n' "$head" "$grad" "$mark" > "$work/full.svg"
# Maskable for Android: the mark kept inside the 80% safe circle.
printf '%s%s<rect width="512" height="512" fill="url(#bg)"/><g transform="translate(66.5 66.5) scale(0.74)">%s</g></svg>\n' "$head" "$grad" "$mark" > "$work/maskable.svg"

for size in 16 32 48; do rsvg-convert -w $size -h $size "$out/icon.svg" -o "$work/$size.png"; done
magick "$work/16.png" "$work/32.png" "$work/48.png" "$out/favicon.ico"
rsvg-convert -w 192 -h 192 "$out/icon.svg" -o "$out/icon-192.png"
rsvg-convert -w 512 -h 512 "$out/icon.svg" -o "$out/icon-512.png"
rsvg-convert -w 512 -h 512 "$work/maskable.svg" -o "$out/icon-maskable-512.png"
rsvg-convert -w 180 -h 180 "$work/full.svg" -o "$out/apple-touch-icon.png"
echo "Icons written to $out"
