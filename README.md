# Uniclip - Universal Clipboard

Thanks to [Aaryan](https://github.com/aaryanporwal) for the idea!

# What is Uniclip?

Apple users, did you know you could copy from one device and paste on the other? Wouldn't it be awesome if you could do that for non-Apple devices too?

Now you can, Apple device or not!

You don't even have to sign in like you need to on Apple devices. You don't have to install Go either!

# Installing

## macOS

```sh
brew install quackduck/tap/uniclip
```
or
```sh
curl -sSL https://github.com/quackduck/uniclip/blob/master/dist/uniclip_darwin_amd64/uniclip\?raw=true > /usr/local/bin/uniclip
chmod +x /usr/local/bin/uniclip
```

## GNU/Linux

*Note: xclip or xsel is needed for uniclip to work on GNU/Linux*

```sh
brew install quackduck/tap/uniclip
```
or
```sh
curl -sSL https://github.com/quackduck/uniclip/blob/master/dist/uniclip_linux_amd64/uniclip\?raw=true -o /usr/local/bin/uniclip # you might need to use sudo
chmod +x /usr/local/bin/uniclip
```

## Windows

Just grab a precompiled binary from this [directory](dist)

# Usage

Run this to start a new clipboard:

 ```sh
uniclip
```

Example output:

```
Starting a new clipboard!
Run `uniclip 192.168.86.24:51607` to join this clipboard

```

Just enter what it says (`uniclip 192.168.86.24:51607`) on your other device with Uniclip installed and hit enter. That's it! Now you can copy from one device and paste on the other.

You can even have multiple devices joined to the same clipboard (Just run that same command on the new device).

```
Uniclip - Universal Clipboard
With Uniclip, you can copy from one device and paste on another.

Usage: uniclip [--verbose/-v] [ <address> | --help/-h ]
Examples:
   uniclip                          # start a new clipboard
   uniclip 192.168.86.24:53701      # join the clipboard at the address - 192.168.86.24:53701
   uniclip --help                   # print this help message
   uniclip -v 192.168.86.24:53701   # enable verbose output
Running just `uniclip` will start a new clipboard.
It will also provide an address with which you can connect to the same clipboard with another device.
```

**Note: The devices have to be on the same local network (eg. connected to the same WiFi).**
