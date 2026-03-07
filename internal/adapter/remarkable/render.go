package remarkable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func SavePageToFile(dir string, pageID string, data []byte) (string, error) {
	path := filepath.Join(dir, pageID+".rm")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", path, err)
	}
	return path, nil
}

func BuildRmcCommand(rmPath string, svgPath string) *exec.Cmd {
	return exec.Command("rmc", "-o", svgPath, rmPath)
}

const remarkableScreenWidth = 1404

func BuildCairoSVGCommand(svgPath string, pngPath string) *exec.Cmd {
	script := fmt.Sprintf(
		"import sys, cairosvg, io; from PIL import Image; "+
			"png = cairosvg.svg2png(url=sys.argv[1], output_width=%d); "+
			"img = Image.open(io.BytesIO(png)); "+
			"bg = Image.new('RGB', img.size, (255,255,255)); "+
			"bg.paste(img, mask=img.split()[3] if img.mode=='RGBA' else None); "+
			"bg.save(sys.argv[2])",
		remarkableScreenWidth)
	return exec.Command("python3", "-c", script, svgPath, pngPath)
}

func RenderPageToPNG(dir string, pageID string, rmData []byte) (string, error) {
	rmPath, err := SavePageToFile(dir, pageID, rmData)
	if err != nil {
		return "", err
	}

	svgPath := filepath.Join(dir, pageID+".svg")
	cmd := BuildRmcCommand(rmPath, svgPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("rmc failed: %w\n%s", err, out)
	}

	pngPath := filepath.Join(dir, pageID+".png")
	cmd = BuildCairoSVGCommand(svgPath, pngPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("cairosvg failed: %w\n%s", err, out)
	}

	return pngPath, nil
}
