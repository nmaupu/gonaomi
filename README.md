# GoNaomi

Command line tool to run Sega Naomi games written in Go.

# Usage

## Standalone mode

```
./bin/gonaomi-darwin -a <naomi_ip> send -f <path_to_game.bin>
```


## Server mode

```
./bin/gonaomi-darwin -a 192.168.12.36 server -l 1337 -r <naomi_roms_path>
```

The server mode runs a REST api ready to get requests.

Available endpoints on the API:
- `/list`: list all available games
- `/load/<game_name>`: load a specific game

> Due to how Naomi board handles network, the server makes sure that only one upload is possible at a time.
> When a request is done to load a game from the API, further call to `/load` will be blocked until ending of the previous call.
> Naomi network often crashes or hangs so a *timeout* has been set to *30 seconds* to be able to unlock the server process in case of problems. If that happens, it is sometimes necessary to reboot the Naomi board.

# Building

Choose between:
```
make linux
make darwin
make freebsd
```

# Credits

This work is heavily based on *triforce tool scripts* and some other codes.
Huge thanks to :
- Debugmode's triforce tools script
- https://github.com/JanekvO/naomi_netboot_upload
