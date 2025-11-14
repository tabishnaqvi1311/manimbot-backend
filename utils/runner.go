package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func RunCode(code string) (string, int, error) {
	tempDir, err := os.MkdirTemp("", "manim-")
	if err != nil {
		return "", 0, err
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "animation.py")
	if err := os.WriteFile(tempFile, []byte(code), 0644); err != nil {
		return "", 0, err
	}

	outputFile := fmt.Sprintf("animation_%d.mp4", time.Now().Unix())

	cmd := exec.Command(
		"manim",
		"-ql",
		"--media_dir", "static",
		tempFile,
		"Scene",
		"-o", outputFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", 0, fmt.Errorf("manim execution failed: %v\nOutput: %s", err, string(output))
	}

	dir, _ := os.Getwd()
	staticDir := filepath.Join(dir, "static", "videos", "animation")

	var videoFullPath string
	var videoRelativePath string

	qualityDirs := []string{"480p15", "720p30", "1080p60"}
	for _, quality := range qualityDirs {
		possiblePath := filepath.Join(staticDir, quality, outputFile)
		if _, err := os.Stat(possiblePath); err == nil {
			videoFullPath = possiblePath
			videoRelativePath = "/static/videos/animation/" + quality + "/" + outputFile
			break
		}
	}

	if videoFullPath == "" {
		return "", 0, fmt.Errorf("could not find generated video file")
	}

	duration, err := GetVideoDuration(videoFullPath)
	if err != nil {
		fmt.Printf("Warning: Could not get video duration: %v\n", err)
		duration = 60
	}

	return videoRelativePath, duration, nil
}

func GetVideoDuration(videoPath string) (int, error) {

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("video file does not exist: %s", videoPath)
	}

	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe execution failed: %v", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse duration: %v", err)
	}

	return int(duration + 0.5), nil
}
