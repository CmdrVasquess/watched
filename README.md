# Watch Elite: Dangerous Journal Files

This is a Go library to watch Journal files of
[Elite: Dangerous](https://www.elitedangerous.com/). The
[commander's Journal](https://forums.frontier.co.uk/forums/elite-api-and-tools/)
is official way to hook player tools into E:D.

This root package only defines some abstractions for event processing. This
gives developers the opportunity to build different processing strategies
for the events coming from ED.

**Building**

Using `watched` as a library does not need anything special. Just

```
import "github.com/CmdrVasquess/watched"
```

in your application.

If you want to build `edeh` and use it right in place, you have to build the
project. Want to get platform independent, instead of `make` try to use

```
$ go run mk/mk.go
```
