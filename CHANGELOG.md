# Changelog

All notable changes to this project will be documented in this file.
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
- **Stress Test:** HTTP Flood module.
- **Cleanup:** Self-destruct (`uninstall`) command.
