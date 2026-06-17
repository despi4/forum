package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	log.Println("checking Go toolchain...")

	if err := ensureGo(); err != nil {
		log.Fatal(err)
	}

	cc, err := ensureCompiler()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("using compiler: %s\n", cc)

	if err := runProject(cc); err != nil {
		log.Fatal(err)
	}
}

func ensureGo() error {
	path, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go not found in PATH")
	}

	cmd := exec.Command(path, "version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go found but failed to start: %w", err)
	}

	return nil
}

func ensureCompiler() (string, error) {
	// 1. Сначала ищем в PATH
	if path, err := exec.LookPath("gcc"); err == nil {
		if err := checkCompiler(path); err == nil {
			return "gcc", nil
		}
	}

	if path, err := exec.LookPath("clang"); err == nil {
		if err := checkCompiler(path); err == nil {
			return "clang", nil
		}
	}

	// 2. Для Windows дополнительно ищем gcc в типичных путях MSYS2
	if runtime.GOOS == "windows" {
		if gccPath, err := findWindowsGCC(); err == nil {
			fmt.Println("gcc found outside PATH:", gccPath)
			return gccPath, nil
		}
	}

	fmt.Println("C compiler not found.")
	fmt.Println("go-sqlite3 requires CGO and a C compiler.")
	fmt.Print("Show installation instructions now? [y/N]: ")

	if !askYes() {
		return "", errors.New("compiler is required, installation cancelled by user")
	}

	if err := installCompiler(); err != nil {
		return "", err
	}

	return "", errors.New("after installing the compiler, restart terminal/IDE and run again")
}

func findWindowsGCC() (string, error) {
	possiblePaths := []string{
		`C:\msys64\ucrt64\bin\gcc.exe`,
		`C:\msys64\mingw64\bin\gcc.exe`,
		`C:\msys64\clang64\bin\clang.exe`,
		`C:\Program Files\MSYS2\ucrt64\bin\gcc.exe`,
		`C:\Program Files\MSYS2\mingw64\bin\gcc.exe`,
		`C:\Program Files\MSYS2\clang64\bin\clang.exe`,
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			if err := checkCompiler(p); err == nil {
				return p, nil
			}
		}
	}

	return "", errors.New("gcc/clang not found in common MSYS2 locations")
}

func checkCompiler(path string) error {
	cmd := exec.Command(path, "--version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func askYes() bool {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))
	return text == "y" || text == "yes"
}

func installCompiler() error {
	switch runtime.GOOS {
	case "windows":
		return installCompilerWindows()
	case "darwin":
		return installCompilerMac()
	case "linux":
		return installCompilerLinux()
	default:
		return fmt.Errorf("unsupported OS for automatic install: %s", runtime.GOOS)
	}
}

func installCompilerWindows() error {
	fmt.Println()
	fmt.Println("Automatic MSYS2 install via winget was removed because it is unstable on some systems.")
	fmt.Println("Please install a compiler manually once, then rerun this command.")
	fmt.Println()
	fmt.Println("Recommended option for Windows:")
	fmt.Println("1. Install MSYS2")
	fmt.Println("2. Open MSYS2 UCRT64 terminal")
	fmt.Println("3. Run:")
	fmt.Println("   pacman -Syu")
	fmt.Println("   pacman -S --needed mingw-w64-ucrt-x86_64-gcc")
	fmt.Println()
	fmt.Println("Then add this folder to PATH:")
	fmt.Println(`C:\msys64\ucrt64\bin`)
	fmt.Println()
	fmt.Println("After that restart terminal or VS Code and run the project again.")

	return errors.New("compiler is not installed")
}

func installCompilerMac() error {
	// First try Apple's command line tools.
	fmt.Println("Trying to install Xcode Command Line Tools...")
	cmd := exec.Command("xcode-select", "--install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err == nil {
		fmt.Println("Xcode Command Line Tools installer launched.")
		fmt.Println("Finish installation, then restart terminal/IDE if needed.")
		return nil
	}

	// If xcode-select did not work, try Homebrew gcc as fallback.
	if _, brewErr := exec.LookPath("brew"); brewErr == nil {
		fmt.Println("Trying to install gcc with Homebrew...")
		brewCmd := exec.Command("brew", "install", "gcc")
		brewCmd.Stdout = os.Stdout
		brewCmd.Stderr = os.Stderr
		brewCmd.Stdin = os.Stdin
		if err := brewCmd.Run(); err != nil {
			return fmt.Errorf("failed to install gcc with brew: %w", err)
		}
		return nil
	}

	return errors.New("could not start Xcode CLT install and brew was not found; install Xcode Command Line Tools manually")
}

func installCompilerLinux() error {
	if _, err := exec.LookPath("apt"); err == nil {
		fmt.Println("Trying to install build-essential with apt...")
		cmd := exec.Command("sudo", "apt", "update")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("apt update failed: %w", err)
		}

		cmd = exec.Command("sudo", "apt", "install", "-y", "build-essential")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("apt install failed: %w", err)
		}
		return nil
	}

	if _, err := exec.LookPath("dnf"); err == nil {
		fmt.Println("Trying to install gcc with dnf...")
		cmd := exec.Command("sudo", "dnf", "install", "-y", "gcc", "gcc-c++", "make")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("dnf install failed: %w", err)
		}
		return nil
	}

	if _, err := exec.LookPath("yum"); err == nil {
		fmt.Println("Trying to install gcc with yum...")
		cmd := exec.Command("sudo", "yum", "install", "-y", "gcc", "gcc-c++", "make")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("yum install failed: %w", err)
		}
		return nil
	}

	if _, err := exec.LookPath("pacman"); err == nil {
		fmt.Println("Trying to install base-devel with pacman...")
		cmd := exec.Command("sudo", "pacman", "-Sy", "--noconfirm", "base-devel", "gcc")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pacman install failed: %w", err)
		}
		return nil
	}

	return errors.New("no supported package manager found; install gcc manually")
}

func runProject(cc string) error {
	cmd := exec.Command("go", "run", "./cmd/api")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	msysBin := `C:\msys64\ucrt64\bin`
	path := os.Getenv("PATH")

	env := os.Environ()
	env = append(env,
		"CGO_ENABLED=1",
		"CC="+cc,
		"MSYSTEM=UCRT64",
		"PATH="+msysBin+";"+path,
	)
	cmd.Env = env

	return cmd.Run()
}