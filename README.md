# geotracker

A simple tracker for Geoguessr data.

The scoring is relatively simple, the higher your Score, the more accurate you can
guess that country.
100 bonus points are there if you check the correct country checkbox.

Example:
- Andorra: 5000 (this would mean you always score perfect on Andorra)
- Japan: 3381 (this would mean you score an average of 3381 points, +- 100 points)
- Australia: 981 (this would mean, you score about 981 points, learn your regions ;D)

## Installation

Should be in the future found here: (TODO!)

## Build
On NixOS you can just: `nix develop` and then call `fyne-build-linux`

On macOS: TODO

On Windows: `go build --release` should be enough, in case it doesn't work, isn't install [fyne v2](https://docs.fyne.io/started/)
