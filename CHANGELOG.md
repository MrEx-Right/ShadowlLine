# Changelog

All notable changes to this project will be documented in this file.

## [1.1.0] - 2026-02-20
### 👻 The "Ghost Protocol" Update

### 🔒 Security
- **End-to-End Encryption (E2EE):** All Agent-to-C2 communication (heartbeats and task results) is now fully encrypted using **AES-256-GCM**. The framework no longer transmits plain JSON payloads, successfully neutralizing static network analysis, DPI, and basic IDS/IPS signatures.

### ✨ Added
- **Remote Self-Update Mechanism:** Introduced the highly requested `update <URL>` command. Agents can now silently download a new binary from a hosted URL, gracefully replace their current running executable on disk (using a `.old` backup method), and restart without dropping the C2 session.
- **Deterministic Agent Fingerprinting:** Agent IDs are no longer randomized on every startup. Implemented an MD5-based persistent fingerprinting system (`Hostname + Username`) to ensure unique and stable identities across system reboots and agent crashes.
- **Smart Reconnection Alerts:** The C2 server now tracks agent states dynamically. It silently logs regular heartbeats to keep the operator's console clean, but triggers a `[+] AGENT RECONNECTED` visual alert if an agent returns after being offline or dead for >15 seconds.

### 🐛 Fixed
- **Exfiltration Bug (404 Error):** Fixed a critical routing bug where the C2 HTTP listener lacked the `/upload` endpoint, which previously caused the `screenshot` and `upload` modules to fail.
- **Duplicate/Zombie Agents:** Resolved an issue where a single restarting agent would spam the database with multiple duplicate entries due to randomized ID generation.
- **Console Spoil:** Fixed the UI flooding issue in Shadow-Shell. Heartbeats are now processed strictly in the background unless a significant state change occurs.

## [1.0.0] - 2026-02-18

### 🎉 Initial Release
First public release of **ShadowLine**, a modern, stealthy, and modular C2 framework designed for Red Team operations.

### ✨ Key Features
- **Cross-Platform Agents:** Full support for Windows, Linux, and macOS targets.
- **Resilient Infrastructure:** Implemented "Dead Drop" resolution via **GitHub Gists** and encrypted tunneling via **Ngrok**.
- **Stealth:** Added **Ghost Mode** (hidden window) and **Fake Error Message** for social engineering on Windows.
- **Persistence:** Registry (Windows) and Crontab (Linux/Mac) integration for auto-start.

### ⚔️ Modules Added
- **Command:** Remote Shell execution.
- **File I/O:** Upload and Download capabilities.
- **Surveillance:** Screenshot capture.
