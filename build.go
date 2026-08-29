//go:build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dotGauge          = ".gauge"
	plugins           = "plugins"
	GOARCH            = "GOARCH"
	goOS              = "GOOS"
	x86               = "386"
	x86_64            = "amd64"
	DARWIN            = "darwin"
	LINUX             = "linux"
	WINDOWS           = "windows"
	bin               = "bin"
	newDirPermissions = 0755
	gauge             = "gauge"
	flash             = "flash-server"
	deploy            = "deploy"
	pluginJSONFile    = "plugin.json"
	webDir            = "web"
	cgoEnabled        = "CGO_ENABLED"
)

var deployDir = filepath.Join(deploy, "flash")

func main() {
	flag.Parse()
	if *install {
		updatePluginInstallPrefix()
		installPlugin(*pluginInstallPrefix)
	} else if *distro {
		createPluginDistro(*allPlatforms)
	} else {
		compile()
	}
}

func compile() {
	buildFrontend()
	copyDist()
	if *allPlatforms {
		compileAcrossPlatforms()
	} else {
		compileGoPackage()
	}
}

func npmInstall() {
	if _, err := os.Stat(filepath.Join(webDir, "node_modules")); err == nil {
		return
	}
	log.Println("Installing npm dependencies...")
	install := "install"
	if _, err := os.Stat(filepath.Join(webDir, "package-lock.json")); err == nil {
		install = "ci"
	}
	cmd := exec.Command("npm", install)
	cmd.Dir = webDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("Failed to install frontend dependencies: %s", err))
	}
}

// buildFrontend 构建 Vue 前端
func buildFrontend() {
	npmInstall()
	log.Println("Building Vue frontend...")
	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = webDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("Failed to build frontend: %s", err))
	}
	log.Println("Frontend built successfully.")
}

// copyDist 复制前端构建产物到根目录 dist（供 embed 使用）
func copyDist() {
	log.Println("Copying dist...")
	src := filepath.Join(webDir, "dist")
	dst := "dist"
	os.RemoveAll(dst)
	if err := mirrorDir(src, dst); err != nil {
		panic(fmt.Sprintf("Failed to copy dist: %s", err))
	}
	log.Println("Dist copied successfully.")
}

// createPluginDistro 创建 Gauge 插件分发包
func createPluginDistro(forAllPlatforms bool) {
	if forAllPlatforms {
		for _, platformEnv := range platformEnvs {
			setEnv(platformEnv)
			*binDir = filepath.Join(bin, fmt.Sprintf("%s_%s", platformEnv[goOS], platformEnv[GOARCH]))
			fmt.Printf("Creating distro for platform => OS:%s ARCH:%s \n", platformEnv[goOS], platformEnv[GOARCH])
			createDistro()
		}
	} else {
		createDistro()
	}
	log.Printf("Distributables created in directory => %s \n", deploy)
}

func createDistro() {
	packageName := fmt.Sprintf("flash-%s-%s-%s", getPluginVersion(), getGOOS(), getArch())
	distroDir := filepath.Join(deploy, packageName)
	copyPluginFiles(distroDir)
	createZip(deploy, packageName)
	os.RemoveAll(distroDir)
}

// createZip 跨平台压缩（Windows 用 PowerShell，其他用 zip）
func createZip(dir, name string) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	zipFile := filepath.Join(wd, dir, name+".zip")
	sourceDir := filepath.Join(wd, dir, name)

	if runtime.GOOS == WINDOWS {
		// Windows 使用 PowerShell
		psCmd := fmt.Sprintf(
			`Compress-Archive -Path "%s\*" -DestinationPath "%s" -Force`,
			sourceDir, zipFile,
		)
		cmd := exec.Command("powershell", "-Command", psCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			panic(fmt.Sprintf("Failed to zip: %s", err))
		}
	} else {
		// Linux/Mac 使用 zip
		os.Chdir(sourceDir)
		output, err := executeCommand("zip", "-r", zipFile, ".")
		fmt.Println(output)
		if err != nil {
			panic(fmt.Sprintf("Failed to zip: %s", err))
		}
		os.Chdir(wd)
	}

	log.Printf("Created: %s.zip\n", filepath.Join(dir, name))
}

func isExecMode(mode os.FileMode) bool {
	return (mode & 0111) != 0
}

func mirrorFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if sfi.Mode()&os.ModeType != 0 {
		log.Fatalf("mirrorFile can't deal with non-regular file %s", src)
	}
	dfi, err := os.Stat(dst)
	if err == nil &&
		isExecMode(sfi.Mode()) == isExecMode(dfi.Mode()) &&
		(dfi.Mode()&os.ModeType == 0) &&
		dfi.Size() == sfi.Size() &&
		dfi.ModTime().Unix() == sfi.ModTime().Unix() {
		return nil
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, newDirPermissions); err != nil {
		return err
	}

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	n, err := io.Copy(df, sf)
	if err == nil && n != sfi.Size() {
		err = fmt.Errorf("copied wrong size for %s -> %s: copied %d; want %d", src, dst, n, sfi.Size())
	}
	cerr := df.Close()
	if err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(dst, sfi.Mode())
	}
	if err == nil {
		err = os.Chtimes(dst, sfi.ModTime(), sfi.ModTime())
	}
	return err
}

func mirrorDir(src, dst string) error {
	log.Printf("Copying '%s' -> '%s'\n", src, dst)
	return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		suffix, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("Failed to find Rel(%q, %q): %v", src, path, err)
		}
		return mirrorFile(path, filepath.Join(dst, suffix))
	})
}

func runProcess(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Execute %v\n", cmd.Args)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func executeCommand(command string, arg ...string) (string, error) {
	cmd := exec.Command(command, arg...)
	bytes, err := cmd.Output()
	return strings.TrimSpace(string(bytes)), err
}

func compileGoPackage() {
	os.Setenv(cgoEnabled, "0")
	runProcess("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", getExecutablePath(flash))
}

func getExecutablePath(file string) string {
	return filepath.Join(getBinDir(), getExecutableName(file))
}

func getExecutableName(file string) string {
	if getGOOS() == WINDOWS {
		return file + ".exe"
	}
	return file
}

func getBinDir() string {
	if *binDir != "" {
		return *binDir
	}
	return filepath.Join(bin, fmt.Sprintf("%s_%s", getGOOS(), getGOARCH()))
}

func copyFiles(files map[string]string, installDir string) {
	for src, dst := range files {
		base := filepath.Base(src)
		installDst := filepath.Join(installDir, dst)
		log.Printf("Copying %s -> %s\n", src, installDst)
		stat, err := os.Stat(src)
		if err != nil {
			panic(err)
		}
		if stat.IsDir() {
			err = mirrorDir(src, installDst)
		} else {
			err = mirrorFile(src, filepath.Join(installDst, base))
		}
		if err != nil {
			panic(err)
		}
	}
}

func copyPluginFiles(destDir string) {
	files := make(map[string]string)
	if getGOOS() == WINDOWS {
		files[filepath.Join(getBinDir(), flash+".exe")] = bin
	} else {
		files[filepath.Join(getBinDir(), flash)] = bin
	}
	files[pluginJSONFile] = ""
	copyFiles(files, destDir)
}

func getPluginVersion() string {
	props, err := getPluginProperties(pluginJSONFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to get properties file. %s", err))
	}
	return props["version"].(string)
}

func setEnv(envVariables map[string]string) {
	for k, v := range envVariables {
		os.Setenv(k, v)
	}
}

var install = flag.Bool("install", false, "Install to the specified prefix")
var pluginInstallPrefix = flag.String("plugin-prefix", "", "Specifies the prefix where the plugin will be installed")
var distro = flag.Bool("distro", false, "Creates distributables for the plugin")
var allPlatforms = flag.Bool("all-platforms", false, "Compiles or creates distributables for all platforms")
var binDir = flag.String("bin-dir", "", "Specifies OS_PLATFORM specific binaries to install when cross compiling")

var platformEnvs = []map[string]string{
	{GOARCH: x86_64, goOS: DARWIN},
	{GOARCH: "arm64", goOS: DARWIN},
	{GOARCH: x86, goOS: LINUX},
	{GOARCH: x86_64, goOS: LINUX},
	{GOARCH: "arm64", goOS: LINUX},
	{GOARCH: x86, goOS: WINDOWS},
	{GOARCH: x86_64, goOS: WINDOWS},
	{GOARCH: "arm64", goOS: WINDOWS},
}

func getPluginProperties(jsonPropertiesFile string) (map[string]interface{}, error) {
	data, err := os.ReadFile(jsonPropertiesFile)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func compileAcrossPlatforms() {
	for _, platformEnv := range platformEnvs {
		setEnv(platformEnv)
		fmt.Printf("Compiling for platform => OS:%s ARCH:%s \n", platformEnv[goOS], platformEnv[GOARCH])
		compileGoPackage()
	}
}

func installPlugin(installPrefix string) {
	copyPluginFiles(deployDir)
	pluginInstallPath := filepath.Join(installPrefix, "flash", getPluginVersion())
	mirrorDir(deployDir, pluginInstallPath)
}

func updatePluginInstallPrefix() {
	if *pluginInstallPrefix == "" {
		if runtime.GOOS == WINDOWS {
			*pluginInstallPrefix = os.Getenv("APPDATA")
			if *pluginInstallPrefix == "" {
				panic(fmt.Errorf("Failed to find AppData directory"))
			}
			*pluginInstallPrefix = filepath.Join(*pluginInstallPrefix, gauge, plugins)
		} else {
			userHome := os.Getenv("HOME")
			if userHome == "" {
				panic(fmt.Errorf("Failed to find User Home directory"))
			}
			*pluginInstallPrefix = filepath.Join(userHome, dotGauge, plugins)
		}
	}
}

func getArch() string {
	switch getGOARCH() {
	case x86:
		return "x86"
	case "arm64":
		return "arm64"
	default:
		return "x86_64"
	}
}

func getGOARCH() string {
	if v := os.Getenv(GOARCH); v != "" {
		return v
	}
	return runtime.GOARCH
}

func getGOOS() string {
	if v := os.Getenv(goOS); v != "" {
		return v
	}
	return runtime.GOOS
}
