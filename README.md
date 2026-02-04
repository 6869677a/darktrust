# DarkTrust //> Android MITM Manager

DarkTrust is a Magisk-based Android MITM manager designed to make installing, tracking, and managing custom CA certificates reliable on modern Android devices. It provides accurate, real-world status detection for rooted devices and avoids false negatives common with legacy CA checks.

Built for operators who want confidence that their MITM environment is actually active — not just installed.

/////>Features

/////> Smart CA detection
Detects whether a CA is already installed and active
Skips unnecessary installs on startup

/////> Magisk module–based CA injection
Systemless CA installation
Survives reboots
Clean removal via Magisk Module delete

/////> Accurate module state tracking
Installed — module directory exists (supports modules and modules_update)
Registered — Magisk recognizes the module
Active — CA is mounted via Magisk overlay (runtime‑accurate)

/////> HTTP proxy toggle
Enable or disable system proxy instantly

/////> ADB reverse management
One‑click port forwarding for interception setups

/////> Multi‑device support
Automatically detects connected devices
Interactive selector when multiple devices are present

/////> Operator‑friendly UI
Colorized checklist
Real‑time status refresh
Clear warnings and reboot notices

/////> Tested Environment

📱 Device: Pixel 8a
🔓 Root: Magisk
🧪 Root method: RootAVD project
🤖 Android: Modern Android builds compatible with Magisk v24+

Requirements:

/////> ADB installed and accessible in PATH
Rooted Android device with Magisk
USB debugging enabled
Go 1.20+ (for building from source)

/////> Installation
1. Clone the repository
git clone https://github.com/yourusername/darktrust.git
cd darktrust

2. Add your CA certificate(s)
Place your CA file(s) into the certs/ directory:

certs/
 └── ca.der

The tool will automatically detect certificates placed here.

Usage
Run DarkTrust
go run main.go


or build it:

go build -o darktrust
./darktrust

What happens on startup

/////>Detects connected device(s)
Verifies root access
Checks Magisk module install/registration state
Determines whether the CA is already active
Displays a real‑time MITM checklist

/////>Menu Options
Install CA — installs CA via Magisk module (if not already active)
Remove CA — removes Magisk module (reboot required)
Toggle Proxy — enables/disables system HTTP proxy
Toggle adb reverse — enables/disables port forwarding
Refresh — re‑detects system state

How CA Detection Works
DarkTrust does not rely on /system/etc/security/cacerts directly.

/////> Modern Magisk:
Mounts certificates via overlays
Does not physically copy certs into /system

/////> DarkTrust instead:
Verifies Magisk module registration
Confirms overlay presence
Assumes CA is active if Magisk reports the module as loaded

This avoids false ❌ states while remaining accurate.

/////> Security Notice

TLS interception is powerful and dangerous if misused.
Use only on devices you own or are authorized to test
Intercepted traffic may include sensitive data
You are responsible for legal and ethical use

/////> License
MIT License

Author
w0rmer
