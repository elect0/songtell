package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
)

var defaults = map[string]string{
	"player_name":        "No player",
	"art":                "",
	"title":              "No title found",
	"artist":             "No artist found",
	"album":              "No album found",
	"duration_formatted": "00:00",
	"volume":             "Unknown",
	"position":           "0",
	"duration":           "0",
	"url":                "No URL found",
	"status":             "No status",
	"loop":               "Unknown",
	"shuffle":            "Unknown",
	"user":               "No user",
}

type MediaInfo struct {
	PlayerName        string
	ArtURL            string
	Title             string
	Artist            string
	Album             string
	DurationFormatted string
	Volume            string
	Position          int64 // Microseconds
	Duration          int64 // Microseconds
	URL               string
	Status            string
	LoopStatus        string
	ShuffleStatus     string
	CurrentUser       string
	AudioBackend      string
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func formatDuration(microsecnonds int64) string {
	seconds := microsecnonds / 1_000_000
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func getPlayerctlMetadata(format string) string {
	output, err := runCommand("playerctl", "metadata", "--format", format)
	if err != nil {
		return ""
	}
	return output
}

func getBackend() string {
	if _, err := exec.LookPath("pipewire"); err == nil {
		if cmd := exec.Command("pgrep", "-x", "pipewire"); cmd.Run() == nil {
			return "PipeWire"
		}
	}

	if _, err := exec.LookPath("pulseaudio"); err == nil {
		if cmd := exec.Command("pgrep", "-x", "pulseaudio"); cmd.Run() == nil {
			return "PulseAudio"
		}
	}
	return "ALSA"
}

func getMediaInfo() MediaInfo {
	info := MediaInfo{}

	info.AudioBackend = getBackend()

	info.PlayerName = getPlayerctlMetadata("{{ playerName }}")
	if info.PlayerName == "" {
		info.PlayerName = defaults["player_name"]
	}

	info.ArtURL = getPlayerctlMetadata("{{ mpris:artUrl }}")
	if info.ArtURL == "" {
		info.ArtURL = defaults["art"]
	}

	info.Title = getPlayerctlMetadata("{{ trunc(title, 33) }}")
	if info.Title == "" {
		info.Title = defaults["title"]
	}

	info.Artist = getPlayerctlMetadata("{{ trunc(artist, 32) }}")
	if info.Artist == "" {
		info.Artist = defaults["artist"]
	}

	info.Album = getPlayerctlMetadata("{{ trunc(album, 33) }}")
	if info.Album == "" {
		info.Album = defaults["album"]
	}

	rawDuration := getPlayerctlMetadata("{{ mpris:length }}")
	if rawDuration != "" {
		dur, err := strconv.ParseInt(rawDuration, 10, 64)
		if err == nil {
			info.Duration = dur
			info.DurationFormatted = formatDuration(dur)
		} else {
			info.DurationFormatted = defaults["duration_formatted"]
		}
	} else {
		info.DurationFormatted = defaults["duration_formatted"]
	}

	rawPosition := getPlayerctlMetadata("{{ position }}")
	if rawPosition != "" {
		pos, err := strconv.ParseInt(rawPosition, 10, 64)
		if err == nil {
			info.Position = pos
		}
	}

	info.URL = getPlayerctlMetadata("{{ trunc(xesam:url, 35) }}")
	if info.URL == "" {
		info.URL = defaults["url"]
	}

	status, err := runCommand("playerctl", "status")
	if err == nil {
		info.Status = status
	} else {
		info.Status = defaults["status"]
	}

	loop, err := runCommand("playerctl", "loop")
	if err == nil {
		info.LoopStatus = loop
	} else {
		info.LoopStatus = defaults["loop"]
	}

	shuffle, err := runCommand("playerctl", "shuffle")
	if err == nil {
		info.ShuffleStatus = shuffle
	} else {
		info.ShuffleStatus = defaults["shuffle"]
	}

	volume := defaults["volume"]
	switch info.AudioBackend {
	case "PipeWire":
		output, err := runCommand("wpctl", "get-volume", "@DEFAULT_AUDIO_SINK@")
		if err == nil {
			re := regexp.MustCompile(`(\d+\.\d+)`)
			match := re.FindStringSubmatch(output)
			if len(match) > 1 {
				val, _ := strconv.ParseFloat(match[1], 64)
				volume = fmt.Sprintf("%d%%", int(val*100))
			}
		}
	case "PulseAudio":
		output, err := runCommand("pactl", "get-sink-volume", "@DEFAULT_SINK@")
		if err == nil {
			re := regexp.MustCompile(`(\d+)%`)
			match := re.FindStringSubmatch(output)
			if len(match) > 1 {
				volume = fmt.Sprintf("%s%%", match[1])
			}
		}
	case "ALSA":
		output, err := runCommand("amixer", "get", "Master")
		if err == nil {
			re := regexp.MustCompile(`\[(\d+)%\]`) // ALSA often has it in brackets
			match := re.FindStringSubmatch(output)
			if len(match) > 1 {
				volume = fmt.Sprintf("%s%%", match[1])
			}
		}
	}
	info.Volume = volume

	currentUser, err := user.Current()
	if err == nil {
		info.CurrentUser = currentUser.Username
	} else {
		info.CurrentUser = defaults["user"]
	}

	return info
}

func main() {
	info := getMediaInfo()

	imageString, err := Convert(info.ArtURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	normal := ""
	bright := ""
	for i := 0; i < 8; i++ {
		normal += fmt.Sprintf("\033[4%dm   \033[0m", i)

	}

	for i := 0; i < 8; i++ {
		bright += fmt.Sprintf("\033[10%dm   \033[0m", i)
	}

	const reset = "\033[0m"

	const cyan = "\033[32m"

	infoLines := []string{
		fmt.Sprintf("  \033[32m%s\033[0m#\033[32m%s\033[0m \n", info.CurrentUser, info.PlayerName),

		fmt.Sprintf("\n"),

		fmt.Sprintf("  %s%s \u25c6 %s%s\n", cyan, strings.Repeat("\u2500", 10), strings.Repeat("\u2500", 10), reset),
		fmt.Sprintf("  \n"),
		fmt.Sprintf("  \033[32mstatus\033[0m ~ %s\n", strings.ToLower(info.Status)),
		fmt.Sprintf("\n"),

		fmt.Sprintf("  %s%s \u25c6 %s%s\n", cyan, strings.Repeat("\u2500", 10), strings.Repeat("\u2500", 10), reset),

		fmt.Sprintf("\n"),
		fmt.Sprintf("  \033[32mtitle\033[0m ~ %s\n", info.Title),
		fmt.Sprintf("  \033[32malbum\033[0m ~ %s\n", info.Album),
		fmt.Sprintf("  \033[32martist\033[0m ~ %s\n", info.Artist),
		fmt.Sprintf("  \033[32mduration\033[0m ~ %s / %s (\033[32m%d%s\033[0m)\n", formatDuration(info.Position), info.DurationFormatted, int(float64(info.Position)/float64(info.Duration)*100.0), "%"),
		fmt.Sprintf("\n"),

		fmt.Sprintf("  %s%s \u25c6 %s%s\n", cyan, strings.Repeat("\u2500", 10), strings.Repeat("\u2500", 10), reset),

		fmt.Sprintf("\n"),
		fmt.Sprintf("  \033[32mvol\033[0m ~ %s\n", info.Volume),
		fmt.Sprintf("  \033[32mloop\033[0m ~ %s\n", strings.ToLower(info.LoopStatus)),
		fmt.Sprintf("  \033[32mshuffle\033[0m ~ %s\n", info.ShuffleStatus),
		fmt.Sprintf("  \033[32martwork\033[0m ~ %s\n", info.ArtURL),

		fmt.Sprintf("\n"),

		fmt.Sprintf("  %s%s \u25c6 %s%s\n", cyan, strings.Repeat("\u2500", 10), strings.Repeat("\u2500", 10), reset),

		fmt.Sprintf("\n"),
		fmt.Sprintf("  \033[32maudio\033[0m ~ %s\n", strings.ToLower(info.AudioBackend)),

		fmt.Sprintf("\n"),
		fmt.Sprintf("  %s\n", bright),
		fmt.Sprintf("  %s\n", normal),
	}

	for i, v := range imageString {
		if i >= len(infoLines) {
			fmt.Printf("%s\n", v)
		} else {
			fmt.Printf("%s\t%s", v, infoLines[i])
		}
	}

}
