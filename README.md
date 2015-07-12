# woozy

Dead simple yr.no weather cli tool, hacked together for the hay harvest of 2015

![Screenshot](http://gummiboll.github.io/woozy/woozy.png)

## Setup
1. `go get github.com/gummiboll/woozy`
2. Run woozy, a ~/.woozy cfg file will be created you.
3. Edit above mentioned file with your location.
4. Run woozy again

## Notes:
* Woozy honors yr.no's nextupdate-time. If you need to clear the cache, run it with --cache-clear.
* All weather data comes from http://yr.no
