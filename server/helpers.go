package server

func (src ApplicationSource) RepositoryURL() string {
	if src.GitProvider == GitProviderGithub {
		return "https://github.com/"+src.RepositoryUsername+"/"+src.RepositoryName+".git"
	}
	if src.GitProvider == GitProviderGitlab {
		return "https://gitlab.com/"+src.RepositoryUsername+"/"+src.RepositoryName+".git"
	}
	return ""
}