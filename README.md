# Uniclip - Universal Clipboard

Apple users, did you know you could copy from one device and paste on the other? Wouldn't it be awesome if you could do that for non-Apple devices too?

Now you can, Apple device or not!

You don't even have to sign in like you need to on Apple devices. You don't have to install Go either!

## Usage

Run this to start a new clipboard:

 ```sh
uniclip
```

Example output:

```text
Starting a new clipboard!
Run `uniclip 192.168.86.24:51607` to join this clipboard

```

Just enter what it says (`uniclip 192.168.86.24:51607`) on your other device with Uniclip installed and hit enter. That's it! Now you can copy from one device and paste on the other.

You can even have multiple devices joined to the same clipboard (just run that same command on the new device).

```text
Uniclip - Universal Clipboard
With Uniclip, you can copy from one device and paste on another.

Usage: uniclip [--debug/-d] [ <address> | --help/-h ]
Examples:
   uniclip                          # start a new clipboard
   uniclip 192.168.86.24:53701      # join the clipboard at 192.168.86.24:53701
   uniclip -d                       # start a new clipboard with debug output
   uniclip -d 192.168.86.24:53701   # join the clipboard with debug output
Running just `uniclip` will start a new clipboard.
It will also provide an address with which you can connect to the same clipboard with another device.
```

*Note: The devices have to be on the same local network (eg. connected to the same WiFi) unless the device has a public IP with all ports routed to it. (use the public IP instead of what Uniclip prints in this case)*

## Installing

### macOS

```sh
brew install quackduck/tap/uniclip
```
or

Get an executable from [releases](https://github.com/quackduck/uniclip/releases) and install to `/usr/bin/uniclip`

### GNU/Linux

*Note: At least one of xsel, xclip or wayland is needed for Uniclip to work on GNU/Linux*

```sh
brew install quackduck/tap/uniclip
```
or

Get an executable from [releases](https://github.com/quackduck/uniclip/releases) and install to `/usr/bin/uniclip`

### Android

Get an executable from [releases](https://github.com/quackduck/uniclip/releases) and install to `$PREFIX/usr/bin/uniclip`

Install the Termux app and Termux:API app from the Play Store.
Then, install the Termux:API package from the command line (in Termux) using:
```sh
pkg install termux-api
```
### Windows

Just grab a precompiled binary from [releases](https://github.com/quackduck/uniclip/releases)

## Uninstalling
Uninstalling Uniclip is very easy. If you used a package manager, use its uninstall feature. If not, just delete the Uniclip binary:

On macOS or GNU/Linux, delete `/usr/local/bin/uniclip`  
On Windows, delete it from where you installed it  
On Termux, delete it from `$PREFIX/usr/bin/uniclip`

## Any other business
Have a question, idea or just want to share something? Head over to [Discussions](https://github.com/quackduck/uniclip/discussions)

Thanks to [Aaryan](https://github.com/aaryanporwal) for the idea!
