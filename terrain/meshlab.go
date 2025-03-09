package terrain

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RenderPLYToIsometricPNG renders a .ply file in isometric view using MeshLab
// and saves the result as a PNG image.
//
// Parameters:
//   - inputPath: Path to the input .ply file
//   - outputPath: Path where the output PNG will be saved
//
// Returns:
//   - error: Any error that occurred during the rendering process
func CreateIsometricView(inputPath, outputPath string) error {
	// Verify input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Verify input file is a .ply file
	if !strings.HasSuffix(strings.ToLower(inputPath), ".ply") {
		return fmt.Errorf("input file must be a .ply file: %s", inputPath)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create a temporary MeshLab script file for the isometric view settings
	scriptPath := filepath.Join(os.TempDir(), "meshlab_isometric_script.mlx")
	scriptContent := `<!DOCTYPE FilterScript>
<FilterScript>
 <filter name="Transform: Rotate">
  <Param value="45" description="Rotation Angle" type="RichDynamicFloat" name="angle" tooltip="Angle of rotation (in degrees)"/>
  <Param value="1;0;0;0" description="Custom axis" type="RichMatrix44f" name="rotAxis" tooltip="The rotation axis is specified by euler angles"/>
  <Param value="0" description="Rotate Center" type="RichEnum" name="rotCenter" tooltip="Choose a method" enum_val0="origin" enum_val1="barycenter" enum_val2="cursor" enum_val3="custom point"/>
  <Param value="0;0;0" description="Custom Point" type="RichPoint3f" name="customAxis" tooltip="This rotation axis is used only if the 'custom point' option is chosen."/>
 </filter>
 <filter name="Transform: Rotate">
  <Param value="35.264" description="Rotation Angle" type="RichDynamicFloat" name="angle" tooltip="Angle of rotation (in degrees)"/>
  <Param value="0;0;1;0" description="Custom axis" type="RichMatrix44f" name="rotAxis" tooltip="The rotation axis is specified by euler angles"/>
  <Param value="0" description="Rotate Center" type="RichEnum" name="rotCenter" tooltip="Choose a method" enum_val0="origin" enum_val1="barycenter" enum_val2="cursor" enum_val3="custom point"/>
  <Param value="0;0;0" description="Custom Point" type="RichPoint3f" name="customAxis" tooltip="This rotation axis is used only if the 'custom point' option is chosen."/>
 </filter>
</FilterScript>`

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("failed to create MeshLab script: %v", err)
	}
	defer os.Remove(scriptPath)

	// Prepare MeshLab command
	// meshlabserver -i input.ply -o output.png -s script.mlx
	cmd := exec.Command("meshlabserver",
		"-i", inputPath,
		"-o", outputPath,
		"-s", scriptPath,
		"-om", "vn",
	)

	// Capture standard output and error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("meshlabserver execution failed: %v\nOutput: %s", err, string(output))
	}

	// Check if output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("output file was not created: %s\nMeshLab output: %s", outputPath, string(output))
	}

	return nil
}
