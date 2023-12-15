## Documentation
We are working on the development and documentation of the project. We target to release the documentation soon to make it easy for contributors

**GraphQL Documentation** - https://graphql.docs.swiftwave.org/

---

## Development Setup
> **Note** : You need to be on linux or mac to setup the project locally. Windows is not supported yet.

**Prerequisites**

|  Name | Version | Installation |
| --- | --- | --- |
| Git | Latest | [Download & Install](https://git-scm.com/downloads) |
| Golang | Latest | [Download & Install](https://golang.org/doc/install) |
| NodeJS | v18.0 atleast | [Install NodeJS](https://deb.nodesource.com/) |
| Docker | Latest | [Follow this docs to install](https://docs.docker.com/engine/install/) |

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
5. Open a root terminal in correct directory `sudo su`
6. Generate SwiftWave default configuration
   ```bash
   go run . init
   ```
7. Prepare SwiftWave environment
   ```bash
   go run . setup
   ```
8. Start Local Postgres Database
   ```bash
   go run . postgres start
   ```
9. Start HAProxy Service
   ```bash
   go run . haproxy start
   ```
10. Start SwiftWave
   ```bash
   go run . start
   ```
11. Swiftwave Service will be available on `http://localhost:3333`

#### Access GrqphQL Playground
Go to `http://localhost:3333/playground`