# lyricsapi
Go package and server app for retrieving time-stamped lyrics from Spotify.

## Usage

### Use as package

```bash
go get github.com/raitonoberu/lyricsapi
```

```go
package main

import (
	"fmt"
	"time"

	"github.com/raitonoberu/lyricsapi/lyrics"
)

func main() {
	api := lyrics.NewLyricsApi("<YOUR COOKIE>")
	lyrics, err := api.GetByName("Rick Astley Never Gonna Give You Up")
	// Alternatively, specify the track id:
	// lyrics, err := api.Get("4uLU6hMCjMI75M1A2tKUQC")
	if err != nil {
		panic(err)
	}
	if lyrics == nil {
		fmt.Println("Not found")
		return
	}
	for _, line := range lyrics.Lyrics.Lines {
		t := time.UnixMilli(line.Time).Format("04:05")
		fmt.Println(t, line.Words)
	}

}
```
<details>

```
00:19 We're no strangers to love
00:23 You know the rules and so do I
00:28 A full commitment's what I'm thinking of
00:32 You wouldn't get this from any other guy
00:36 I just wanna tell you how I'm feeling
00:41 Gotta make you understand
00:44 Never gonna give you up
00:46 Never gonna let you down
00:48 Never gonna run around and desert you
00:52 Never gonna make you cry
00:54 Never gonna say goodbye
00:56 Never gonna tell a lie and hurt you
01:02 We've known each other for so long
01:06 Your heart's been aching but you're too shy to say it
01:10 Inside we both know what's been going on
01:14 We know the game and we're gonna play it
01:18 And if you ask me how I'm feeling
01:23 Don't tell me you're too blind to see
01:26 Never gonna give you up
01:28 Never gonna let you down
01:30 Never gonna run around and desert you
01:34 Never gonna make you cry
01:36 Never gonna say goodbye
01:39 Never gonna tell a lie and hurt you
01:43 Never gonna give you up
01:45 Never gonna let you down
01:47 Never gonna run around and desert you
01:51 Never gonna make you cry
01:53 Never gonna say goodbye
01:56 Never gonna tell a lie and hurt you
02:01 
02:03 (Give you up)
02:07 ♪
02:09 (Ooh) Never gonna give, never gonna give
02:11 (Give you up)
02:14 ♪
02:18 We've known each other for so long
02:22 Your heart's been aching but you're too shy to say it
02:26 Inside we both know what's been going on
02:30 We know the game and we're gonna play it
02:34 I just wanna tell you how I'm feeling
02:39 Gotta make you understand
02:42 Never gonna give you up
02:44 Never gonna let you down
02:46 Never gonna run around and desert you
02:51 Never gonna make you cry
02:53 Never gonna say goodbye
02:55 Never gonna tell a lie and hurt you
02:59 Never gonna give you up
03:01 Never gonna let you down
03:03 Never gonna run around and desert you
03:08 Never gonna make you cry
03:10 Never gonna say goodbye
03:12 Never gonna tell a lie and hurt you
03:16 Never gonna give you up
03:18 Never gonna let you down
03:20 Never gonna run around and desert you
03:25 Never gonna make you cry
03:27 Never gonna say goodbye
03:28 
```
</details>

### Deploy to Vercel

1. https://vercel.com/docs/get-started
2. Set "COOKIE" enviroment variable.

### Where do I get cookie?

Press F12, open the `Network` tab and go to [open.spotify.com](https://open.spotify.com/). Click on the first request to `open.spotify.com`. Scroll down to the `Request Headers`, right click the `cookie` field and select `Copy value`.

## License

The Unlicense, see [LICENSE](./LICENSE) for additional information.
