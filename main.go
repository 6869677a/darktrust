package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

/* ---------- COLORS ---------- */
const (
	R = "\033[31m"
	G = "\033[32m"
	Y = "\033[33m"
	B = "\033[34m"
	C = "\033[36m"
	X = "\033[0m"
)

/* ---------- CONFIG ---------- */
const (
	moduleID   = "darktrust"
	modulePath = "/data/adb/modules/darktrust"
	certDir    = "certs"
)

/* ---------- STATE ---------- */
type State struct {
	Device     string
	HasRoot    bool
	Installed  bool // module folder exists
	Registered bool // Magisk sees it
	Active     bool // CA mounted by Magisk
	ProxyOn    bool
	ReverseOn  bool
	CertFiles  []string
	CALoaded   bool // smart startup flag
}

var state State

/* ---------- UTILS ---------- */
func die(msg string) {
	fmt.Println(R + msg + X)
	os.Exit(1)
}

func run(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func adb(args ...string) (string, error) {
	all := append([]string{"adb", "-s", state.Device}, args...)
	return run(all...)
}

func su(cmd string) (string, error) {
	return adb("shell", "su", "-mm", "-c", "sh -c \""+cmd+"\"")
}

func confirm(msg string) bool {
	fmt.Print(Y + msg + " [y/N]: " + X)
	var c string
	fmt.Scanln(&c)
	return strings.ToLower(c) == "y"
}

/* ---------- BANNER ---------- */
func banner() {
	fmt.Print(R + `
██████╗  █████╗ ██████╗ ██╗  ██╗████████╗██████╗ ██╗   ██╗███████╗████████╗
██╔══██╗██╔══██╗██╔══██╗██║ ██╔╝╚══██╔══╝██╔══██╗██║   ██║██╔════╝╚══██╔══╝
██║  ██║███████║██████╔╝█████╔╝    ██║   ██████╔╝██║   ██║███████╗   ██║
██║  ██║██╔══██║██╔══██╗██╔═██╗    ██║   ██╔══██╗██║   ██║╚════██║   ██║
██████╔╝██║  ██║██║  ██║██║  ██╗   ██║   ██║  ██║╚██████╔╝███████║   ██║
╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚══════╝   ╚═╝
` + X)
	fmt.Println(C + "DarkTrust — Android MITM Manager (FINAL RC Modern Magisk)\n" + X)
}

/* ---------- DEVICE ---------- */
func pickDevice() {
	out, _ := run("adb", "devices")
	lines := strings.Split(out, "\n")
	var devs []string
	for _, l := range lines {
		if strings.HasSuffix(l, "\tdevice") {
			devs = append(devs, strings.Fields(l)[0])
		}
	}
	if len(devs) == 0 {
		die("No adb devices detected")
	}
	if len(devs) == 1 {
		state.Device = devs[0]
		return
	}

	fmt.Println("Select device:")
	for i, d := range devs {
		fmt.Printf(" %d) %s\n", i+1, d)
	}
	fmt.Print("> ")
	var c int
	fmt.Scanln(&c)
	state.Device = devs[c-1]
}

/* ---------- DETECTION ---------- */
func detectRoot() {
	out, _ := adb("shell", "su", "-mm", "-c", "id")
	state.HasRoot = strings.Contains(out, "uid=0")
}

func detectModule() {
	// Installed = folder exists in either path
	out, _ := adb("shell", "test -d /data/adb/modules/darktrust || test -d /data/adb/modules_update/darktrust && echo yes || echo no")
	state.Installed = strings.TrimSpace(out) == "yes"

	// Registered = magisk recognizes module
	out, _ = su("magisk --list-modules | grep " + moduleID)
	state.Registered = strings.TrimSpace(out) != ""

	// Active = assume true if registered
	state.Active = state.Registered

	// Smart startup flag
	state.CALoaded = state.Active
}

func detectProxy() {
	out, _ := adb("shell", "settings", "get", "global", "http_proxy")
	state.ProxyOn = strings.Contains(out, "8080") && !strings.Contains(out, ":0")
}

func detectReverse() {
	out, _ := run("adb", "reverse", "--list")
	state.ReverseOn = strings.Contains(out, "tcp:8080")
}

/* ---------- CERTS ---------- */
func findCerts() {
	state.CertFiles = nil
	entries, _ := os.ReadDir(certDir)
	for _, e := range entries {
		if !e.IsDir() {
			state.CertFiles = append(state.CertFiles, certDir+"/"+e.Name())
		}
	}
}

/* ---------- ACTIONS ---------- */
func installCerts() {
	if state.CALoaded {
		fmt.Println(G + "✓ CA already active in system, no install needed" + X)
		return
	}

	if !state.HasRoot {
		die("Root not granted")
	}
	findCerts()
	if len(state.CertFiles) == 0 {
		die("No certs in ./certs/")
	}

	fmt.Println(Y + "Installing CA(s) via Magisk module..." + X)

	var sh strings.Builder
	sh.WriteString("rm -rf " + modulePath + "\n")
	sh.WriteString("mkdir -p " + modulePath + "/system/etc/security/cacerts\n")

	sh.WriteString("cat <<EOF > " + modulePath + "/module.prop\n")
	sh.WriteString("id=" + moduleID + "\n")
	sh.WriteString("name=DarkTrust CA\n")
	sh.WriteString("version=1.0\n")
	sh.WriteString("versionCode=1\n")
	sh.WriteString("author=w0rmer\n")
	sh.WriteString("description=Injected system CA(s)\n")
	sh.WriteString("EOF\n")

	for i, cert := range state.CertFiles {
		tmp := fmt.Sprintf("/data/local/tmp/ca.%d", i)
		dst := fmt.Sprintf("%s/system/etc/security/cacerts/ca.%d", modulePath, i)

		if out, err := adb("push", cert, tmp); err != nil {
			fmt.Println(out)
			die("adb push failed")
		}

		sh.WriteString("cp " + tmp + " " + dst + "\n")
		sh.WriteString("chmod 644 " + dst + "\n")
	}

	out, err := su(sh.String())
	if err != nil {
		fmt.Println(out)
		die("Install failed")
	}

	verify, _ := su("ls " + modulePath + "/system/etc/security/cacerts | wc -l")
	fmt.Println(G+"✓ Installed CA count:"+X, verify)

	if confirm("Reboot now to activate CA(s)?") {
		adb("reboot")
	}
}

func removeCerts() {
	if !confirm("Remove DarkTrust CA module?") {
		return
	}
	out, err := su("rm -rf " + modulePath + " /data/local/tmp/ca.*")
	if err != nil {
		fmt.Println(out)
		die("Removal failed")
	}
	fmt.Println(G + "✓ CA module removed" + X)
	if confirm("Reboot now to finalize removal?") {
		adb("reboot")
	}
}

/* ---------- TOGGLES ---------- */
func toggleProxy() {
	if state.ProxyOn {
		adb("shell", "settings", "put", "global", "http_proxy", ":0")
		fmt.Println(Y + "HTTP proxy disabled" + X)
	} else {
		adb("shell", "settings", "put", "global", "http_proxy", "127.0.0.1:8080")
		fmt.Println(G + "HTTP proxy enabled" + X)
	}
}

func toggleReverse() {
	if state.ReverseOn {
		run("adb", "reverse", "--remove", "tcp:8080")
		fmt.Println(Y + "ADB reverse removed" + X)
	} else {
		run("adb", "reverse", "tcp:8080", "tcp:8080")
		fmt.Println(G + "ADB reverse created" + X)
	}
}

/* ---------- MENU ---------- */
func mark(ok bool) string {
	if ok {
		return G + "✓" + X
	}
	return R + "✗" + X
}

func checklist() {
	fmt.Println(B + "MITM CHECKLIST\n" + X)
	fmt.Println(" Device:", state.Device)
	fmt.Println(" Root:", mark(state.HasRoot))
	fmt.Println(" Magisk CA module installed:", mark(state.Installed))
	fmt.Println(" Magisk CA module registered:", mark(state.Registered))
	fmt.Println(" CA active in system:", mark(state.Active))
	fmt.Println(" HTTP proxy:", mark(state.ProxyOn))
	fmt.Println(" adb reverse:", mark(state.ReverseOn))
	fmt.Println(Y + "\n TLS unpinning = operator responsibility\n" + X)
}

func menu() {
	r := bufio.NewReader(os.Stdin)
	for {
		detectRoot()
		detectModule()
		detectProxy()
		detectReverse()
		findCerts()

		banner()
		checklist()

		fmt.Println("1) Install CA(s)")
		fmt.Println("2) Remove CA(s)")
		fmt.Println("3) Toggle Proxy")
		fmt.Println("4) Toggle adb reverse")
		fmt.Println("5) Refresh")
		fmt.Println("q) Quit")
		fmt.Print("> ")

		c, _ := r.ReadString('\n')
		switch strings.TrimSpace(c) {
		case "1":
			installCerts()
		case "2":
			removeCerts()
		case "3":
			toggleProxy()
		case "4":
			toggleReverse()
		case "5":
		case "q":
			return
		}

		fmt.Print("\nPress Enter...")
		r.ReadString('\n')
	}
}

/* ---------- MAIN ---------- */
func main() {
	pickDevice()
	detectModule() // smart startup detect
	if state.CALoaded {
		fmt.Println(G + "✓ CA already active in system, no install needed" + X)
	}
	menu()
}
