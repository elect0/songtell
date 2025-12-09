# songtell
A simple, light-weight, fastfetch-like Go CLI tool that quickly displays information about the song currently playing. Currently, it only supports Linux.

<img src="https://i.postimg.cc/bvprgHVb/2025-12-09-191417-hyprshot.png" />

## instalation
At this moment, Songtell isn't available on AUR. You have to manually install it.

1. Download from latest release:
```bash
wget https://github.com/elect0/songtell/releases/download/v1.0.0/songtell-linux-amd64.tar.gz
```

2. Extract the archive:
```bash
tar -xzf songtell_linux_amd64.tar.gz
```

3. Move the binary to a folder in your path:
```bash
sudo mv songtell /usr/local/bin/
sudo chmod +x /usr/local/bin/songtell
```

4. Verify instalation:
```bash
songtell
```

Songtell doesn’t require any prerequisites or dependencies—it’s written in Go.

## Player Compatibility

Works with most MPRIS2-compatible players out of the box:

- Spotify, VLC, Firefox, Chrome
- Rhythmbox, Clementine, Strawberry

**Terminal/daemon** players may need an MPRIS bridge (mpdris2, cmus, moc-mpris) and playerctld enabled via systemd.

## Notes
- Album artwork respects terminal color schemes.
- Does **not work with youtube** because artwork ratios are incompatible.

## Acknowledgment
This project is inspired by/enhanced from https://github.com/ekrlstd/songfetch.

