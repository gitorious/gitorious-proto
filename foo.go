
func execute(handler ProtocolHandler, internalApi InternalApi, logger Logger) (responseText, error) {
	username, err := handler.GetUsername()
	if err != nil {
		logger.Printf("%v, aborting...", err)
		return "...", ERROR ,401
	}

	repoPath, command, err := handler.ParseCommand()
	if err != nil {
		logger.Printf("%v, aborting...", err)
		return "Invalid command", ERROR ,400
	}

	repoConfig, err := internalApi.GetRepositoryConfiguration(repoPath, username)
	if err != nil {
		logger.Printf("%v, aborting...", err)
		return "Access denied or invalid repository path", ERROR, 403
	}

	// fullRepoPath := filepath.Join(reposRootPath, repoConfig.RealPath)
	// gitShellCommand := formatGitShellCommand(command, fullRepoPath)
	// handler.Printf(`invoking git-shell with command "%v"`, gitShellCommand)

	if stderr, err := handler.RunProxy(repoConfig, command, username); err != nil {
		logger.Printf("error occured in git-shell: %v", err)
		logger.Printf("stderr: %v", stderr)
		return "Fatal error, please contact support", ERROR, 500
	}

	logger.Printf("done")
	return "", nil
}
