package generate

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/harish876/hypefx/internal/cli/commands"
	"github.com/harish876/hypefx/internal/utils"
	"github.com/spf13/cobra"
)

var (
	BASE_PATH = "github.com/harish876/hypefx/internal/cli/commands/generate/scaffolding"
)

var GenerateCmd = &cobra.Command{
	Use:     "generate [project_name/module_name](optional)",
	Short:   "Generates a new HypeFX Project Structure",
	Long:    `Generates a new HypeFX Project Structure, when a base path to the project i.e the go mod base path is provided.`,
	Example: "hype generate",
	Run:     generate,
}

//go:embed scaffolding
var embeddedFiles embed.FS

func generate(cmd *cobra.Command, args []string) {
	var message string
	var errorInterface error
	var moduleName interface{}

	config, err := commands.GetConfig()
	if err != nil {
		slog.Error("generate", "GetAllConfig func call", err)
		errorInterface = errors.Join(err, fmt.Errorf("unable to get config file"))
		DisplayError(errorInterface)
		return
	}
	if len(args) >= 1 {
		moduleName = args[0]
		config.Module = moduleName.(string)
		if err := commands.SetConfig(config); err != nil {
			DisplayError(err)
			return
		}
	} else {
		modName := config.Module
		slog.Debug("generate", "moduleName", modName)
		if modName == "" && len(args) == 0 {
			errorInterface = fmt.Errorf("unable to find module name. use hype generate [module_name] to automagically perform the initializations")
			DisplayError(errorInterface)
			return
		}
		moduleName = modName
	}
	slog.Debug("generate", "moduleName", moduleName)
	initCmd := exec.Command("go", "mod", "init", moduleName.(string))
	_, err = initCmd.CombinedOutput()
	if err != nil {
		errorInterface = fmt.Errorf("error initializing Go module.\ngo mod might already exist or there might be no go mod file.run go mod init [module_name] manually.\ncheck error with key 'go mod init error' at tmp/hypefx.log for more info")
		DisplayError(errorInterface)
		slog.Debug("generate", "go mod init error", err)
		return
	}
	if err := copyDirectory("scaffolding", ".", moduleName.(string)); err != nil {
		errorInterface = fmt.Errorf("unable to scaffold project")
		DisplayError(errors.Join(errorInterface, fmt.Errorf("check logs with key 'scaffolding error' at tmp/hypefx for more info")))
		slog.Debug("generate", "scaffolding error", err)
		return
	}
	slog.Debug("generate", "file config", config)

	basePath, _ := os.Getwd()
	if config.AppDir == "" {
		path := filepath.Join(basePath, "app")
		slog.Debug("generate", "set appDir", path)
		config.AppDir = path
	}
	if config.RoutesDir == "" {
		path := filepath.Join(basePath, "routes")
		slog.Debug("generate", "set routesDir", path)
		config.RoutesDir = path
	}
	if !config.Routing {
		slog.Debug("generate", "set routing", true)
		config.Routing = true
	}
	if err := commands.SetConfig(config); err != nil {
		DisplayError(err)
		return
	}
	message = fmt.Sprintf("Successfully Instantiated a Hype FX Project in module %s\n\n1.Run `go mod tidy`\n2.Run `npm install` or `npm i`\n4.Start the dev server using `npm run dev`\n4.😅 no more of this npm business.custom build command coming soon...\n5.And.. you can find CLI logs at tmp/hypefx.log", moduleName.(string))
	DisplayMessage(message)
}

func copyDirectory(src, dst, moduleName string) error {
	files, err := embeddedFiles.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		sourcePath := filepath.Join(src, file.Name())
		destPath := filepath.Join(dst, file.Name())

		slog.Debug("copyDirectory", "file name", file.Name(), "file type", file.Type().IsRegular())

		if file.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			if err := copyDirectory(sourcePath, destPath, moduleName); err != nil {
				return err
			}
		} else {
			if err := copyFile(sourcePath, destPath, moduleName); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst, moduleName string) error {
	sourceFile, err := embeddedFiles.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	utils.ReplaceFileContent(dst, BASE_PATH, moduleName)
	return nil
}
