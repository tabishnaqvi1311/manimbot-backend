package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func RunCode(code string) (string, error) {
	tempDir, err := os.MkdirTemp("", "manim-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "animation.py")
	if err := os.WriteFile(tempFile, []byte(code), 0644); err != nil {
		return "", err
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

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("manim execution code failed")
	}

	return "/static/videos/animation/480p15/" + outputFile, nil

}