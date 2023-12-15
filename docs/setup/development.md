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
1. Create a new user using CLI
   ```bash
   go run . create-user
   ```
2. Use Login Endpoint for generating a JWT Token. **Ref** - [REST Api Documentation](https://github.com/swiftwave-org/swiftwave/blob/develop/docs/rest_api.md)
3. Go to `http://localhost:3333/playground`
4. In headers section, add authorization details
   ```json
   {
     "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDI0OTc4MDMsImlhdCI6MTcwMjQ5NDIwMywibmJmIjoxNzAyNDk0MjAzLCJ1c2VybmFtZSI6InRhbm1veXNydCJ9.9Bx_Og9FzG09Wi-TjNndzO7U1yLZURT1itmz3VxjuV8"
   }
   ```
5. Now, click on `refresh` icon on playground to get schema details and avail the auto-complete feature
6. You can now start querying and mutating data
7. Refer the docs for more information - [API DOCS](https://github.com/swiftwave-org/swiftwave/blob/develop/docs/api_docs.md)
