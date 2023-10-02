# Setup Guide
1. [Development Setup](#development-setup)
2. [Production Setup](#production-setup)
---

## Development Setup
**Prerequisites**
|  Name | Version | Installation |
| --- | --- | --- |
| Git | Latest | [Download & Install](https://git-scm.com/downloads) |
| NodeJS | v18.0 atleast | [Install NodeJS](https://deb.nodesource.com/) |
| Docker | Latest | [Follow this docs to install](https://docs.docker.com/engine/install/) |

> **Note For Windows :** Supports only Hyper-V based docker installation. WSL-2 based docker installation is not supported yet.

### Steps
1. Fork this repository [if you want to contribute]
2. Clone the repository
    ```bash
    git clone git@github.com:<username>/swiftwave.git
    ```
3. Go to the cloned directory
    ```bash
    cd swiftwave
    ```
4. Run `npm install`
5. Run the setup script
    - For  Linux
      ```bash
      ./dev.linux.sh
      ```
    - For macOS
      ```bash
      ./dev.mac.sh
      ```
    - For Windows [PowerShell]
      ```bash
      dev.windows.ps1
      ```
6. Then follow the instructions printed in the terminal after the script execution is completed.

---

## Production Setup
**We have tested the setup on `Ubuntu`, `Debian` and `AWS Linux 2`. We are working toward making the installer compatible with other Linux distros as well.**

`Swiftwave` can be installed with one click.
> We recommend using a fresh server for production setup. If you are using an existing server, please ensure you have stopped all the services running on ports 80 and 443.

Run this command in bash
```bash
 curl -L get.swiftwave.org | bash
```

That's all 🍻

Wait for ⏰  few minutes, and it will become online.
