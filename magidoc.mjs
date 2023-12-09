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
            appLogo: 'https://github.com/swiftwave-org.png'
        }
    }
  }