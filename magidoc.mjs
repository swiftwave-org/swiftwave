function markdown(string) {
  // Takes the first indent and trims that length from everywhere.
  // Markdown templates don't like the extra space at the beginning.
  const target = string[0]
  const trimSize = /^\s+/.exec(string)[0].length
  return target
    .split('\n')
    .map((line) => line.substr(trimSize - 1))
    .join('\n')
}


export default {
    introspection: {
      type: 'sdl',
      paths: ['swiftwave_service/graphql/schema/**/*.graphqls'],
    },
    website: {
        template: 'carbon-multi-page',
        output: './graphql-docs',
        options: {
            appTitle: 'Swiftwave GraphQL Documentation',
            appLogo: 'https://github.com/swiftwave-org.png',
            siteMeta: {
              description: "Documentation for Swiftwave's GraphQL API.",
              'og:description': "Documentation for Swiftwave's GraphQL API.",
            },
            pages: [
              {
                title: 'Welcome',
                content: markdown`
                  # 👋 Hi
      
                  Welcome to the documentation of the Swiftwave GraphQL API.

                  ## Don't know about Swiftwave?

                  SwiftWave is a self-hosted lightweight PaaS solution to deploy and manage your applications on any VPS without any hassle 👀
                
                  ## Want to support us?
                  Star ⭐ our [GitHub repository](https://github.com/swiftwave-org/swiftwave) and join our [Slack community](https://join.slack.com/t/swiftwave-team/shared_invite/zt-21n86aslx-aAvBi3hv1GigVA_XoXiu4Q) to get help and support from our team.
                  `,
              }
            ],
        }
    }
  }