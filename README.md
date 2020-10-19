# Uniclip - Universal Clipboard

Thanks to [Aaryan](https://github.com/aaryanporwal) for the idea!

# What is Uniclip?

Apple users, did you know you could copy from one device and paste on the other? Wouldn't it be awesome if you could do that for non-Apple devices too?

Now you can, Apple device or not!

You don't even have to sign in like you need to on Apple devices. You don't have to install Go either!

# Installing

## macOS

```
brew install quackduck/quackduck/uniclip
```
or
```
curl -sSL https://github.com/quackduck/uniclip/blob/master/dist/uniclip_darwin_amd64/uniclip\?raw=true > /usr/local/bin/uniclip
chmod +x /usr/local/bin/uniclip
```

## Linux

```
brew install quackduck/quackduck/uniclip
```
or
```
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

**Note: The devices have to be on the same local network (eg. connected to the same WiFi)**
