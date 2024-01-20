const fs = require('fs')
const os = require('os');
const path = require('path');
const crypto = require('crypto');
const axios = require('axios').default;
const { execSync } = require('child_process')
const ghpages = require('gh-pages');

/**
 * Working directory structure
 * /tmp/<random_string>
 *     /<branch_name>
 *     /graphql-docs
 */

// Constant variables
const CURRENT_DIRECTORY = process.cwd()
const GITHUB_REPO = 'swiftwave-org/swiftwave'
const GRAPHQL_DOCUMENTATION_BRANCH = 'docs/graphql'
const CURRENT_BRANCH = execSync(`git branch | grep \\* | cut -d ' ' -f2`, { cwd: CURRENT_DIRECTORY, stdio: 'pipe' }).toString().trim()

// Index page for documentation
const INDEX_PAGE = ``


// Build docs for GraphQL
async function buildGraphQlDocs() {
    const WORKING_DIRECTORY = path.join(CURRENT_DIRECTORY, ".build_docs_tmp")

    // Create a working directory in the system's temp folder
    fs.mkdirSync(WORKING_DIRECTORY, { recursive: true })
    console.log(`Created temporary directory: ${WORKING_DIRECTORY}`)

    // Repo URL
    const REPO_URL = `https://codeload.github.com/${GITHUB_REPO}/zip/refs/heads/${GRAPHQL_DOCUMENTATION_BRANCH}`
    console.log(`Repo URL: ${REPO_URL}`)

    // Create the folder to move existing docs
    const GRAPHQL_DOCUMENTATION_FOLDER = path.join(WORKING_DIRECTORY, 'graphql-docs')
    fs.mkdirSync(GRAPHQL_DOCUMENTATION_FOLDER, { recursive: true })
    console.log(`Created folder: ${GRAPHQL_DOCUMENTATION_FOLDER}`)

    // Check if the branch exists in the repo by HEAD request
    let isExistGraphQLDocsBranch = false
    try {
        const response = await axios.head(REPO_URL)
        if (response.status === 200) {
            isExistGraphQLDocsBranch = true
        } else {
            throw new Error('Branch does not exist')
        }
    }
    catch (err) {
        if (err.response.status === 404) {
            isExistGraphQLDocsBranch = false
        } else {
            throw err
        }
    }

    // If the branch exists, download the zip file
    if (isExistGraphQLDocsBranch) {
        const zipFilePath = path.join(WORKING_DIRECTORY, 'graphql-docs.zip')
        const writer = fs.createWriteStream(zipFilePath)

        const response = await axios({
            url: REPO_URL,
            method: 'GET',
            responseType: 'stream'
        })

        response.data.pipe(writer)

        await new Promise((resolve, reject) => {
            writer.on('finish', resolve)
            writer.on('error', reject)
        })

        console.log('Downloaded zip file')

        // Extract the zip file in a tmp folder
        const extract = require('extract-zip')
        const extractPath = path.join(WORKING_DIRECTORY, 'downloaded-zip-content')
        await extract(zipFilePath, { dir: extractPath })
        console.log('Extracted zip file')

        // Move extractPath/swiftwave-<branch_name>/ to WORKING_DIRECTORY/graphql-docs
        const docs_branch_folder = GRAPHQL_DOCUMENTATION_BRANCH.split('/').join('-')
        const extractedFolder = path.join(extractPath, `swiftwave-${docs_branch_folder}`)
        fs.renameSync(extractedFolder, GRAPHQL_DOCUMENTATION_FOLDER)
        console.log('Moved extracted folder to graphql-docs')
    }

    // Generate the GraphQL documentation
    // Run `npx magidoc generate`
    const magidocCommand = 'npx magidoc generate --file ' + generateMjs(CURRENT_BRANCH)
    execSync(magidocCommand, { cwd: CURRENT_DIRECTORY, stdio: 'inherit' })
    console.log('Generated GraphQL documentation')

    // Check if CURRENT_DIRECTORY/graphql-docs exists
    const CURRENT_DIRECTORY_GRAPHQL_DOCUMENTATION_FOLDER = path.join(CURRENT_DIRECTORY, 'graphql-docs')
    const isExistCurrentDirectoryGraphqlDocs = fs.existsSync(CURRENT_DIRECTORY_GRAPHQL_DOCUMENTATION_FOLDER)
    if (isExistCurrentDirectoryGraphqlDocs === false) {
        console.log("Failed to generate GraphQL documentation")
        return
    }

    // Delete if GRAPHQL_DOCUMENTATION_FOLDER/<branch_name> exists
    const GRAPHQL_DOCUMENTATION_CURRENT_BRANCH_FOLDER = path.join(GRAPHQL_DOCUMENTATION_FOLDER, CURRENT_BRANCH)
    const isExistGraphqlDocsBranchFolder = fs.existsSync(GRAPHQL_DOCUMENTATION_CURRENT_BRANCH_FOLDER)
    if (isExistGraphqlDocsBranchFolder) {
        fs.rmSync(GRAPHQL_DOCUMENTATION_CURRENT_BRANCH_FOLDER, { recursive: true })
        console.log(`Deleted ${GRAPHQL_DOCUMENTATION_CURRENT_BRANCH_FOLDER}`)
    }

    // Create the folder GRAPHQL_DOCUMENTATION_CURRENT_BRANCH_FOLDER
    fs.mkdirSync(GRAPHQL_DOCUMENTATION_CURRENT_BRANCH_FOLDER, { recursive: true })

    // Move CURRENT_DIRECTORY/graphql-docs to WORKING_DIRECTORY/graphql-docs/
    execSync(`mv ${CURRENT_DIRECTORY_GRAPHQL_DOCUMENTATION_FOLDER}/* ${GRAPHQL_DOCUMENTATION_CURRENT_BRANCH_FOLDER}`, { stdio: 'inherit' })

    // Copy `graphql.docs.html` to `GRAPHQL_DOCUMENTATION_FOLDER/index.html`
    execSync(`cp ./graphql.docs.html ${GRAPHQL_DOCUMENTATION_FOLDER}/index.html`, { stdio: 'inherit' })

    // Publish the GraphQL documentation
    await ghpages.publish(GRAPHQL_DOCUMENTATION_FOLDER, {
        branch: GRAPHQL_DOCUMENTATION_BRANCH
    })

    console.log('Published GraphQL documentation')

    // Delete the temporary directory [if it exists]
    try {
        fs.accessSync(WORKING_DIRECTORY, fs.constants.F_OK)
        fs.rmSync(WORKING_DIRECTORY, { recursive: true })
        console.log("Temporary folder deleted")
    }
    catch (err) {
        console.log("Temporary folder failed to delete > ", WORKING_DIRECTORY)
    }
}

// Function to generate mjs for specific branch
function generateMjs(branchName) {
    const mjsContent = `export default {
        introspection: {
          type: 'sdl',
          paths: ['swiftwave_service/graphql/schema/**/*.graphqls'],
        },
        website: {
            template: 'carbon-multi-page',
            output: './graphql-docs',
            options: {
                siteRoot: '/${branchName}',
                appTitle: 'Swiftwave GraphQL Documentation',
                appLogo: 'https://github.com/swiftwave-org.png',
                siteMeta: {
                    description: "Documentation for Swiftwave's GraphQL API.",
                    'og:description': "Documentation for Swiftwave's GraphQL API.",
                  },
                  pages: [
                    {
                      title: 'Welcome',
                      content: markdown\`
                        # 👋 Hi
            
                        Welcome to the documentation of the Swiftwave GraphQL API.
      
                        ## Don't know about Swiftwave?
      
                        SwiftWave is a self-hosted lightweight PaaS solution to deploy and manage your applications on any VPS without any hassle 👀
                      
                        ## Want to support us?
                        Star ⭐ our [GitHub repository](https://github.com/swiftwave-org/swiftwave) and join our [Slack community](https://join.slack.com/t/swiftwave-team/shared_invite/zt-21n86aslx-aAvBi3hv1GigVA_XoXiu4Q) to get help and support from our team.
                        \`,
                    }
                  ],
            }
        }
    }`
    const tmpFolder = path.join(os.tmpdir(), crypto.randomBytes(20).toString('hex'))
    fs.mkdirSync(tmpFolder, { recursive: true })
    const mjsFilePath = path.join(tmpFolder, 'magidoc.mjs')
    fs.writeFileSync(mjsFilePath, mjsContent)
    return mjsFilePath
}

buildGraphQlDocs()
    .then(() => console.log('Done!'))

