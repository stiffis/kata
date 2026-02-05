<p align="center">
  <img src="assets/logo.png" alt="KATA Logo" width="200">
</p>

<h1 align="center">🥋 KATA</h1>

<p align="center">
  <strong>"Slow is smooth. Smooth is fast."</strong><br>
  The minimalist, high-performance typing trainer for the terminal.
</p>

<p align="center">
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"></a>
  <a href="https://github.com/charmbracelet/bubbletea"><img src="https://img.shields.io/badge/UI-Charm%20Stack-FF4A00?style=for-the-badge" alt="Charm Stack"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPL--3.0-blue.svg?style=for-the-badge" alt="License"></a>
</p>

---

## What is KATA?

**KATA** is a terminal-based typing trainer inspired by martial arts philosophy. It's not just about typing fast; it's about internalizing movements until code flows from your fingers without conscious thought.

Based on the **Monkeytype** aesthetic and powered by **Go**, KATA is designed for developers who want to master their craft.

## Features

- **Multi-language:** Support for English, Spanish, French, and German.
- **Developer Mode:** Practice with real syntax from **Go, Rust, Python, C++, and JavaScript**.
- **Smart Analysis:** Track WPM, accuracy, and detect your "weak keys."
- **Keyboard Heatmap:** Visualize which keys are giving you the most trouble.
- **Beautiful Themes:** Catppuccin, Nord, Dracula, Rose Pine, and more.
- **File Loading:** Practice with your own text or code files.
- **Zen Mode:** Remove distractions and focus entirely on the text.

## Screenshots

<p align="center">
  <img src="assets/p1.png" width="800" alt="Practice Mode">
  <br>
  <em>Centered practice interface with real-time error highlighting.</em>
</p>

<p align="center">
  <img src="assets/p2.png" width="800" alt="Statistics View">
  <br>
  <em>Detailed statistics and keyboard heatmap.</em>
</p>

## Quick Installation

Ensure you have Go installed and run:

```bash
git clone https://github.com/stiffis/kata.git
cd kata
chmod +x install.sh
./install.sh
```

Now, simply type `kata` in your terminal.

## Usage

- `kata`: Opens the interactive menu.
- `kata --file <path>`: Practice with a specific file.
- `kata --stats`: View your accumulated progress.
- `kata --theme dracula`: Change the theme quickly.

## License

This project is licensed under the **GPL-3.0 License**. See the [LICENSE](LICENSE) file for more details.

---
