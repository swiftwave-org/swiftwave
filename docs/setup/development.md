## Development Setup

**Documentation -** https://github.com/swiftwave-org/swiftwave/blob/develop/docs/api_docs.md

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
    git clone git@github.com:<username>/swiftwave.git --recursive
    ```
3. Go to the cloned directory
    ```bash
    cd swiftwave
    ```
4. Run `npm install`
5. Run `npm run build:dashboard`
6. Initialize Docker Swarm (if not already initialized)
   ```bash
   docker swarm init
   ```
7. Open a root terminal in correct directory `sudo su`
Generate SwiftWave default configuration
   ```bash
   go run . init
   ```
8. Prepare SwiftWave environment
   ```bash
   go run . setup
   ```
9. Disable `TLS` in configuration

   a. Open a root terminal in correct directory `sudo su`
   
   b. Run `EDITOR=nano go run . config` or `EDITOR=vim go run . config`
   
   c. Change `service.use_tls` to `false`
   
   d. Save the file and exit

10. Start Local Postgres Database
   ```bash
   go run . postgres start
   ```
11. Start HAProxy Service
   ```bash
   go run . haproxy start
   ```
12. Start SwiftWave
   ```bash
   go run . start --dev
   ```
13. Swiftwave Service will be available on `http://localhost:3333`

#### Access Swiftwave Dashboard
1. Go to `http://localhost:3333`
2. Login using your credentials
   > If you have not created any user, you can create one using CLI
   ```bash
    go run . create-user
    ```

#### Access GrqphQL Playground
1. Create a new user using CLI
   ```bash
   go run . create-user
   ```
2. Use Login Endpoint for generating a JWT Token. **Ref** - [REST Api Documentation](https://github.com/swiftwave-org/swiftwave/blob/develop/docs/rest_api.md)
   You can also generate the token using curl command
    ```bash
   curl --location 'http://localhost:3333/auth/login' \
   --form 'username="admin"' \
   --form 'password="admin"'
   ```
3. Go to `http://localhost:3333/playground`
4. In headers section, add authorization details
   ```json
   {
     "Authorization": "Bearer <token_retrieved_from_login>"
   }
   ```
5. Now, click on `refresh` icon on playground to get schema details and avail the auto-complete feature
6. You can now start querying and mutating data
7. Refer the docs for more information - [API DOCS](https://github.com/swiftwave-org/swiftwave/blob/develop/docs/api_docs.md)